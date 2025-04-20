//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/server"
)

func TestLoginUser_NotFound_TableDriven(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		role              api.UserRole
		emailPrefix       string
		expectedErrorCode string
		expectedStatus    int
		expectedTitle     string
		expectedDetailFmt string
	}{
		{
			name:              "DoctorNotFound",
			role:              api.UserRoleDoctor,
			emailPrefix:       "not.a.doctor.",
			expectedErrorCode: "doctor.not-found",
			expectedStatus:    http.StatusNotFound,
			expectedTitle:     "Not Found",
		},
		{
			name:              "PatientNotFound",
			role:              api.UserRolePatient,
			emailPrefix:       "not.a.patient.",
			expectedErrorCode: "patient.not-found",
			expectedStatus:    http.StatusNotFound,
			expectedTitle:     "Not Found",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nonExistentEmail := fmt.Sprintf("%s%s@example.com", tc.emailPrefix, uuid.NewString())

			loginReqPayload := api.LoginUserJSONRequestBody{
				Email: types.Email(nonExistentEmail),
				Role:  tc.role,
			}
			loginReqBytes, err := json.Marshal(loginReqPayload)

			require := require.New(t)
			require.NoError(err, "Failed to marshal login request")

			url := fmt.Sprintf("%s/auth/login", ServerUrl)

			res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(loginReqBytes))
			require.NoError(err, "http.Post failed for /login (%s)", tc.name)
			defer res.Body.Close()

			require.Equal(
				tc.expectedStatus,
				res.StatusCode,
				"Expected status code %d for %s",
				tc.expectedStatus,
				tc.name,
			)

			var errorResponse api.ErrorDetail
			err = json.NewDecoder(res.Body).Decode(&errorResponse)
			require.NoError(err, "Failed to decode error response body for %s", tc.name)

			assert := assert.New(t)
			assert.Equal(tc.expectedStatus, errorResponse.Status, "Error response status mismatch")
			assert.Equal(tc.expectedTitle, errorResponse.Title, "Error response title mismatch")
			assert.Equal(tc.expectedErrorCode, errorResponse.Code, "Error response code mismatch")
			assert.Contains(
				errorResponse.Detail,
				nonExistentEmail,
				"Error detail should mention the email %s",
				nonExistentEmail,
			)
		})
	}
}

func TestLoginUser_Patient_OK(t *testing.T) {
	t.Parallel()

	uniqueEmail := fmt.Sprintf("test.login.patient.%s@example.com", uuid.NewString())
	patientRequest := newPatient(uniqueEmail)
	createdPatient := mustCreatePatient(t, patientRequest)
	require.NotEmpty(t, createdPatient.Id, "Setup failed: Created patient ID is empty")

	loginReqPayload := api.LoginUserJSONRequestBody{
		Email: types.Email(uniqueEmail),
		Role:  api.UserRolePatient,
	}
	loginReqBytes, err := json.Marshal(loginReqPayload)
	require.NoError(t, err, "Failed to marshal login request")

	url := fmt.Sprintf("%s/auth/login", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(loginReqBytes))
	require.NoError(t, err, "http.Post failed for /login")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode, "Expected OK status code for patient login")

	var user api.User
	err = json.NewDecoder(res.Body).Decode(&user)
	require.NoError(t, err, "Failed to decode response body for successful patient login")

	userRole, err := user.Discriminator()
	require.NoError(t, err, "Failed to decode discriminator of logged in user")
	require.Equal(t, string(api.UserRolePatient), userRole)

	loggedInUser, err := user.AsPatient()
	require.NoError(t, err, "Failed to decode logged in user as patient")

	assert := assert.New(t)
	assert.Equal(createdPatient.Id, loggedInUser.Id)
	assert.Equal(createdPatient.Email, loggedInUser.Email)
	assert.Equal(createdPatient.FirstName, loggedInUser.FirstName)
	assert.Equal(createdPatient.LastName, loggedInUser.LastName)
	assert.Equal(api.UserRolePatient, loggedInUser.Role)
}

func TestLoginUser_Doctor_OK(t *testing.T) {
	t.Parallel()

	uniqueEmail := fmt.Sprintf("test.login.doctor.%s@example.com", uuid.NewString())
	docRequest := newDoctor(uniqueEmail)
	createdDoctor := mustCreateDoctor(t, docRequest)
	require.NotEmpty(t, createdDoctor.Id, "Setup failed: Created doctor ID is empty")

	loginReqPayload := api.LoginUserJSONRequestBody{
		Email: types.Email(uniqueEmail),
		Role:  api.UserRoleDoctor,
	}
	loginReqBytes, err := json.Marshal(loginReqPayload)
	require.NoError(t, err, "Failed to marshal login request")

	url := fmt.Sprintf("%s/auth/login", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(loginReqBytes))
	require.NoError(t, err, "http.Post failed for /login")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode, "Expected OK status code for doctor login")

	var user api.User
	err = json.NewDecoder(res.Body).Decode(&user)
	require.NoError(t, err, "Failed to decode response body for successful doctor login")

	userRole, err := user.Discriminator()
	require.NoError(t, err, "Failed to decode discriminator of logged in user")
	require.Equal(t, string(api.UserRoleDoctor), userRole)

	loggedInUser, err := user.AsDoctor()
	require.NoError(t, err, "Failed to decode logged in user as doctor")

	assert := assert.New(t)
	assert.Equal(createdDoctor.Id, loggedInUser.Id)
	assert.Equal(createdDoctor.Email, loggedInUser.Email)
	assert.Equal(createdDoctor.FirstName, loggedInUser.FirstName)
	assert.Equal(createdDoctor.LastName, loggedInUser.LastName)
	assert.Equal(api.UserRoleDoctor, loggedInUser.Role)
	require.NotNil(t, loggedInUser.Specialization, "Doctor specialization should not be nil")
	assert.Equal(createdDoctor.Specialization, loggedInUser.Specialization)
}

func TestCreatePatient(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	patient := newPatient("test@one.com")
	req, err := json.Marshal(patient)
	require.NoError(err)

	url := ServerUrl + "/auth/register"
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(req))
	require.NoError(err)

	require.Equal(http.StatusCreated, res.StatusCode, "response: %+v", res)

	var newPatient api.Patient
	err = json.NewDecoder(res.Body).Decode(&newPatient)
	res.Body.Close()
	require.NoError(err, "response: %+v", res)

	assert.Equal(t, patient.Email, newPatient.Email)
	assert.Equal(t, patient.FirstName, newPatient.FirstName)
	assert.Equal(t, patient.LastName, newPatient.LastName)
}

func TestCreateDoctor(t *testing.T) {
	t.Parallel()
	uniqueEmail := fmt.Sprintf("test.doctor.create.%s@example.com", uuid.NewString())
	docRequest := newDoctor(uniqueEmail)

	createdDoctor := mustCreateDoctor(t, docRequest)

	assert := assert.New(t)
	assert.Equal(docRequest.Email, createdDoctor.Email)
	assert.Equal(docRequest.FirstName, createdDoctor.FirstName)
	assert.Equal(docRequest.LastName, createdDoctor.LastName)
	assert.Equal(docRequest.Specialization, createdDoctor.Specialization)
	assert.NotEmpty(createdDoctor.Id, "Created doctor ID should not be empty")
}

func TestCreatePatient_Conflict(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	assert := assert.New(t)

	uniqueEmail := fmt.Sprintf("test.conflict.%s@example.com", uuid.NewString())
	patient1 := newPatient(uniqueEmail)
	_ = mustCreatePatient(t, patient1)

	res2, err := createPatient(patient1) // Use the same patient data
	require.NoError(err, "Second createPatient call failed")
	defer res2.Body.Close()

	require.Equal(http.StatusConflict, res2.StatusCode, "Expected conflict status code")

	var errorResponse api.ErrorDetail
	err = json.NewDecoder(res2.Body).Decode(&errorResponse)
	require.NoError(err, "Failed to decode error response body")

	assert.Equal(http.StatusConflict, errorResponse.Status)
	assert.Equal("Conflict", errorResponse.Title)
	assert.Equal("patient.email-exists", errorResponse.Code)
	assert.Contains(
		errorResponse.Detail,
		string(patient1.Email),
		"Error detail should mention the conflicting email",
	)
}

func TestCreateDoctor_Conflict(t *testing.T) {
	t.Parallel()
	uniqueEmail := fmt.Sprintf("test.doctor.conflict.%s@example.com", uuid.NewString())
	docRequest1 := newDoctor(uniqueEmail)
	_ = mustCreateDoctor(t, docRequest1)

	res2, err := createDoctor(docRequest1)
	require.NoError(t, err, "Second createDoctor call failed")
	defer res2.Body.Close()

	require.Equal(t, http.StatusConflict, res2.StatusCode, "Expected conflict status code")

	var errorResponse api.ErrorDetail
	err = json.NewDecoder(res2.Body).Decode(&errorResponse)
	require.NoError(t, err, "Failed to decode error response body")

	assert := assert.New(t)
	assert.Equal(http.StatusConflict, errorResponse.Status)
	assert.Equal("Conflict", errorResponse.Title)
	assert.Equal("doctor.email-exists", errorResponse.Code)
	assert.Contains(
		errorResponse.Detail,
		string(docRequest1.Email),
		"Error detail should mention the conflicting email",
	)
}
