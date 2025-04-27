package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Nesquiko/aass/common/mongodb"
)

const (
	resourcesCollection    = "resources"
	reservationsCollection = "reservations"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrResourceUnavailable = errors.New("resource is unavailable during the requested time slot")
)

type ResourceType string

const (
	ResourceTypeMedicine  ResourceType = "medicine"
	ResourceTypeFacility  ResourceType = "facility"
	ResourceTypeEquipment ResourceType = "equipment"
)

type Resource struct {
	Id   uuid.UUID    `bson:"_id"  json:"id"`
	Name string       `bson:"name" json:"name"`
	Type ResourceType `bson:"type" json:"type"`
}

type Reservation struct {
	Id            uuid.UUID    `bson:"_id"           json:"id"`
	AppointmentId uuid.UUID    `bson:"appointmentId" json:"appointmentId"` // Link to the Appointment document
	ResourceId    uuid.UUID    `bson:"resourceId"    json:"resourceId"`    // Link to the specific Resource document (Facility, Equipment, etc.)
	ResourceName  string       `bson:"resourceName"  json:"resourceName"`
	ResourceType  ResourceType `bson:"resourceType"  json:"resourceType"`
	StartTime     time.Time    `bson:"startTime"     json:"startTime"`
	EndTime       time.Time    `bson:"endTime"       json:"endTime"`
}

type mongoResourcesDb struct {
	resources    *mongo.Collection
	reservations *mongo.Collection
}

func newMongoResourceDb(ctx context.Context, uri string, db string) (mongoResourcesDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return mongoResourcesDb{}, fmt.Errorf("newMongoResourceDb: %w", err)
	}

	resourcesColl := mongoDb.Collection(resourcesCollection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "type", Value: 1}},
		Options: options.Index().SetName("idx_resource_type"),
	}
	_, err = resourcesColl.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Could not ensure resource index (may already exist)",
			"error",
			err,
		)
	}

	reservationsColl := mongoDb.Collection(reservationsCollection)
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "resourceId", Value: 1},
				{Key: "startTime", Value: 1},
			},
			Options: options.Index().SetName("idx_reservation_resource_time"),
		},
		{
			Keys:    bson.D{{Key: "appointmentId", Value: 1}},
			Options: options.Index().SetName("idx_reservation_appointmentId"),
		},
	}

	for _, idx := range indexModels {
		_, err = reservationsColl.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"Could not ensure reservation index (may already exist)",
				"error",
				err,
			)
		}
	}

	return mongoResourcesDb{resources: resourcesColl, reservations: reservationsColl}, nil
}

func (db mongoResourcesDb) Disconnect(ctx context.Context) error {
	return db.resources.Database().Client().Disconnect(ctx)
}

func (m *mongoResourcesDb) CreateResource(
	ctx context.Context,
	name string,
	typ ResourceType,
) (Resource, error) {
	resource := Resource{
		Id:   uuid.New(),
		Name: name,
		Type: typ,
	}

	_, err := m.resources.InsertOne(ctx, resource)
	if err != nil {
		return Resource{}, fmt.Errorf(
			"CreateResource creating %q failed to insert document: %w",
			typ,
			err,
		)
	}

	return resource, nil
}

func (m *mongoResourcesDb) ResourceById(ctx context.Context, id uuid.UUID) (Resource, error) {
	filter := bson.M{"_id": id}

	var resource Resource
	err := m.resources.FindOne(ctx, filter).Decode(&resource)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Resource{}, fmt.Errorf("ResourceById %s: %w", id, ErrNotFound)
		}
		return Resource{}, fmt.Errorf("ResourceById query failed for %s: %w", id, err)
	}

	return resource, nil
}

func (m *mongoResourcesDb) CreateReservation(
	ctx context.Context,
	appointmentId uuid.UUID,
	resourceId uuid.UUID,
	resourceName string,
	resourceType ResourceType,
	startTime time.Time,
	endTime time.Time,
) (Reservation, error) {
	if err := m.resourceExists(ctx, resourceId); err != nil {
		return Reservation{}, fmt.Errorf("CreateReservation resource check error: %w", err)
	}
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return Reservation{}, fmt.Errorf("CreateReservation: endTime must be after startTime")
	}

	// --- Conflict Check (excluding self) ---
	// This check MUST happen before the upsert to prevent overwriting a valid
	// reservation from another appointment if the timing overlaps.
	conflictFilter := bson.M{
		"resourceId":    resourceId,
		"appointmentId": bson.M{"$ne": appointmentId}, // Exclude the current appointment
		"startTime":     bson.M{"$lt": endTime},
		"endTime":       bson.M{"$gt": startTime},
	}

	count, err := m.reservations.CountDocuments(ctx, conflictFilter)
	if err != nil {
		return Reservation{}, fmt.Errorf("CreateReservation conflict check failed: %w", err)
	}

	if count > 0 {
		// Conflict found with a *different* appointment
		return Reservation{}, fmt.Errorf(
			"%w: resourceId %s from %s to %s (conflict with another appointment)",
			ErrResourceUnavailable,
			resourceId,
			startTime.Format(time.RFC3339),
			endTime.Format(time.RFC3339),
		)
	}

	// --- Upsert Reservation ---
	// No conflict with other appointments found. Proceed to create or update
	// the reservation for *this* specific appointment and resource.

	// Filter to find the specific reservation for this appointment and resource
	upsertFilter := bson.M{
		"appointmentId": appointmentId,
		"resourceId":    resourceId,
	}

	// Define the fields to set on update or initial insert
	updateFields := bson.M{
		"resourceName": resourceName,
		"resourceType": resourceType,
		"startTime":    startTime,
		"endTime":      endTime,
	}

	// Define the complete update operation using $set and $setOnInsert
	// $setOnInsert ensures _id is only generated when the document is first created.
	updateDefinition := bson.M{
		"$set": updateFields,
		"$setOnInsert": bson.M{
			"_id":           uuid.New(),
			"appointmentId": appointmentId,
			"resourceId":    resourceId,
		},
	}

	// Configure options for FindOneAndUpdate:
	// - Upsert(true): Create the document if it doesn't exist.
	// - ReturnDocument(options.After): Return the document state *after* the update/insert.
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result Reservation
	err = m.reservations.FindOneAndUpdate(ctx, upsertFilter, updateDefinition, opts).
		Decode(&result)
	if err != nil {
		return Reservation{}, fmt.Errorf("CreateReservation upsert failed: %w", err)
	}

	return result, nil
}

func (m *mongoResourcesDb) FindAvailableResourcesAtTime(
	ctx context.Context,
	appointmentDate time.Time,
) (struct {
	Medicines  []Resource
	Facilities []Resource
	Equipment  []Resource
}, error,
) {
	result := struct {
		Medicines  []Resource
		Facilities []Resource
		Equipment  []Resource
	}{
		Medicines:  make([]Resource, 0),
		Facilities: make([]Resource, 0),
		Equipment:  make([]Resource, 0),
	}

	// --- Aggregation Pipeline ---
	// 1. $lookup: Join resources with reservations to find conflicting bookings.
	//    - Use a pipeline within $lookup to filter reservations *before* joining.
	//    - Filter condition: Find reservations where the given appointmentDate
	//      falls within the reservation's time slot [startTime, endTime).
	//      (startTime <= appointmentDate < endTime)
	// 2. $match: Keep only those resources where the lookup found *no*
	//    conflicting reservations (i.e., the resulting array is empty).

	pipeline := mongo.Pipeline{
		// Lookup conflicting reservations
		bson.D{
			{Key: "$lookup", Value: bson.M{
				"from": "reservations",                // The collection to join with
				"let":  bson.M{"resource_id": "$_id"}, // Variable for the resource's ID
				"pipeline": mongo.Pipeline{
					// Sub-pipeline: Filter reservations before joining
					bson.D{
						{Key: "$match", Value: bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$eq": []any{"$resourceId", "$$resource_id"},
									}, // Match resource ID
									{
										"$lte": []any{"$startTime", appointmentDate},
									}, // Reservation starts <= appointmentDate
									{
										"$gt": []any{"$endTime", appointmentDate},
									}, // Reservation ends > appointmentDate
								},
							},
						}},
					},
					bson.D{{Key: "$project", Value: bson.M{"_id": 1}}},
					// Limit to 1 conflict, as we only need to know if *any* exist
					bson.D{{Key: "$limit", Value: 1}},
				},
				"as": "conflictingReservations", // Name of the array field to add
			}},
		},
		// Stage 2: Match resources that have NO conflicting reservations
		bson.D{
			{Key: "$match", Value: bson.M{
				"conflictingReservations": bson.M{"$size": 0}, // Keep only if the array is empty
			}},
		},
	}

	cursor, err := m.resources.Aggregate(ctx, pipeline)
	if err != nil {
		return result, fmt.Errorf("FindAvailableResourcesAtTime aggregation failed: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close available resources cursor", "error", cerr.Error())
		}
	}()

	var availableResources []Resource // Temporarily store all results before grouping
	if err = cursor.All(ctx, &availableResources); err != nil {
		return result, fmt.Errorf("FindAvailableResourcesAtTime decode failed: %w", err)
	}

	for _, resource := range availableResources {
		switch resource.Type {
		case ResourceTypeMedicine:
			result.Medicines = append(result.Medicines, resource)
		case ResourceTypeFacility:
			result.Facilities = append(result.Facilities, resource)
		case ResourceTypeEquipment:
			result.Equipment = append(result.Equipment, resource)
		default:
			slog.Warn(
				"Found resource with unknown type",
				"resourceId",
				resource.Id,
				"type",
				resource.Type,
			)
		}
	}

	if err = cursor.Err(); err != nil {
		return result, fmt.Errorf("FindAvailableResourcesAtTime cursor error: %w", err)
	}

	return result, nil
}

func (m *mongoResourcesDb) DeleteReservationsByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) error {
	filter := bson.M{"appointmentId": appointmentId}

	_, err := m.reservations.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("DeleteReservationsByAppointmentId failed: %w", err)
	}

	return nil
}

func (m *mongoResourcesDb) ResourcesByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) ([]Resource, error) {
	reservations, err := m.ReservationsByAppointmentId(ctx, appointmentId)
	if err != nil {
		return nil, fmt.Errorf("ResourcesByAppointmentIdFromReservations: %w", err)
	}

	resourceMap := make(map[uuid.UUID]Resource)
	for _, reservation := range reservations {
		resource := Resource{
			Id:   reservation.ResourceId,
			Name: reservation.ResourceName,
			Type: reservation.ResourceType,
		}
		resourceMap[reservation.ResourceId] = resource
	}

	resources := make([]Resource, 0, len(resourceMap))
	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	return resources, nil
}

func (m *mongoResourcesDb) ReservationsByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) ([]Reservation, error) {
	filter := bson.M{"appointmentId": appointmentId}

	var reservations []Reservation
	cursor, err := m.reservations.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ReservationsByAppointmentId: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close reservations cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &reservations); err != nil {
		return nil, fmt.Errorf("ReservationsByAppointmentId decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("ReservationsByAppointmentId cursor error: %w", err)
	}

	return reservations, nil
}

func (m *mongoResourcesDb) resourceExists(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	count, err := m.resources.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("resourceExists failed resource count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}
