package mongodb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	tUUID       = reflect.TypeOf(uuid.UUID{})
	uuidSubtype = byte(0x04)
)

func ConnectMongo(ctx context.Context, uri string, db string) (*mongo.Database, error) {
	mongoRegistry := bson.NewRegistry()
	mongoRegistry.RegisterTypeEncoder(tUUID, bson.ValueEncoderFunc(uuidEncodeValue))
	mongoRegistry.RegisterTypeDecoder(tUUID, bson.ValueDecoderFunc(uuidDecodeValue))

	opts := options.Client().ApplyURI(uri).SetRegistry(mongoRegistry)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("ConnectMongo: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ConnectMongo failed to ping mongo server: %w", err)
	}

	return client.Database(db), nil
}

func uuidEncodeValue(ec bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "uuidEncodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}
	b := val.Interface().(uuid.UUID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

func uuidDecodeValue(dc bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{
			Name:     "uuidDecodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		err = vr.ReadNull()
	case bson.TypeUndefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uuid2))
	return nil
}
