package pkg

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/aass/doctor-service/api"
)

// AvailableDoctors implements api.ServerInterface.
func (s DoctorServer) AvailableDoctors(
	w http.ResponseWriter,
	r *http.Request,
	params api.AvailableDoctorsParams,
) {
	panic("unimplemented - isn't used on FE, won't implement")
}

// DoctorsCalendar implements api.ServerInterface.
func (s DoctorServer) DoctorsCalendar(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsCalendarParams,
) {
	calendar, err := s.app.DoctorsCalendar(r.Context(), doctorId, params.From, params.To)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DoctorsCalendar")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, calendar)
}

// DoctorsTimeslots implements api.ServerInterface.
func (s DoctorServer) DoctorsTimeslots(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsTimeslotsParams,
) {
	slots, err := s.app.DoctorTimeSlots(r.Context(), doctorId, params.Date.Time)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DoctorsCalendar")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, slots)
}

// GetDoctorByEmail implements api.ServerInterface.
func (s DoctorServer) GetDoctorByEmail(
	w http.ResponseWriter,
	r *http.Request,
	email api.DoctorEmail,
) {
	patient, err := s.app.DoctorByEmail(r.Context(), string(email))
	if errors.Is(err, ErrNotFound) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "doctor.not-found",
				Title:  "Not Found",
				Detail: fmt.Sprintf("patient with email %q not found.", email),
				Status: http.StatusNotFound,
			},
		}
		encodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetDoctorByEmail", "role", "patient")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}

// GetDoctors implements api.ServerInterface.
func (s DoctorServer) GetDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := s.app.GetAllDoctors(r.Context())
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetAllDoctors")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, api.Doctors{Doctors: doctors})
}

func (s DoctorServer) CreateDoctor(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewDoctorRequest](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	doc, err := s.app.CreateDoctor(r.Context(), req)
	if errors.Is(err, ErrDuplicateEmail) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "doctor.email-exists",
				Title:  "Conflict",
				Detail: fmt.Sprintf("A doctor with email %q already exists.", req.Email),
				Status: http.StatusConflict,
			},
		}
		encodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "CreateDoctor", "role", "doctor")
		encodeError(w, internalServerError())
		return
	}
	encode(w, http.StatusCreated, doc)
	return
}

func (s DoctorServer) GetDoctorById(w http.ResponseWriter, r *http.Request, doctorId api.DoctorId) {
	doctor, err := s.app.DoctorById(r.Context(), doctorId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "doctor.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Doctor with ID %q not found.", doctorId),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		}
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetDoctorById")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, doctor)
}
