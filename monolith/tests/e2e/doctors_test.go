//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/server"
)

func TestGetDoctorById_OK(t *testing.T) {
	t.Parallel()

	docRequest := newDoctor(fmt.Sprintf("test.doctor.get.%s@example.com", uuid.NewString()))
	createdDoctor := mustCreateDoctor(t, docRequest)
	require.NotEmpty(t, createdDoctor.Id, "Setup failed: Created doctor ID is empty")

	url := fmt.Sprintf("%s/doctors/%s", ServerUrl, createdDoctor.Id)
	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for GetDoctorById")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected OK status code")

	var retrievedDoctor api.Doctor
	err = json.NewDecoder(res.Body).Decode(&retrievedDoctor)
	require.NoError(t, err, "Failed to decode response body for GetDoctorById")

	assert := assert.New(t)
	assert.Equal(createdDoctor.Id, retrievedDoctor.Id)
	assert.Equal(createdDoctor.Email, retrievedDoctor.Email)
	assert.Equal(createdDoctor.FirstName, retrievedDoctor.FirstName)
	assert.Equal(createdDoctor.LastName, retrievedDoctor.LastName)
	assert.Equal(createdDoctor.Specialization, retrievedDoctor.Specialization)
}

func TestGetDoctorById_NotFound(t *testing.T) {
	t.Parallel()

	nonExistentID := uuid.New()
	url := fmt.Sprintf("%s/doctors/%s", ServerUrl, nonExistentID)
	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for GetDoctorById (NotFound)")
	defer res.Body.Close()

	require.Equal(t, http.StatusNotFound, res.StatusCode, "Expected Not Found status code")

	var errorResponse api.ErrorDetail
	err = json.NewDecoder(res.Body).Decode(&errorResponse)
	require.NoError(t, err, "Failed to decode error response body for GetDoctorById (NotFound)")

	assert := assert.New(t)
	assert.Equal(http.StatusNotFound, errorResponse.Status)
	assert.Equal("Not Found", errorResponse.Title)
	assert.Equal("doctor.not-found", errorResponse.Code)
	assert.Contains(
		errorResponse.Detail,
		nonExistentID.String(),
		"Error detail should mention the missing ID",
	)
}

func mustCreateDoctor(t *testing.T, request *api.DoctorRegistration) api.Doctor {
	t.Helper()
	require := require.New(t)

	res, err := createDoctor(request)
	require.NoError(err, "mustCreateDoctor: failed during http post")
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	require.NoError(readErr, "mustCreateDoctor: failed to read response body")
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(
		http.StatusCreated,
		res.StatusCode,
		"mustCreateDoctor: unexpected status code. Response body: %s",
		string(bodyBytes),
	)

	var createdDoctor api.Doctor
	err = json.NewDecoder(res.Body).Decode(&createdDoctor)
	require.NoError(err, "mustCreateDoctor: failed to decode successful response")

	return createdDoctor
}

func createDoctor(request *api.DoctorRegistration) (*http.Response, error) {
	if request == nil {
		request = newDoctor("")
	}
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("createDoctor marshal: %w", err)
	}

	url := ServerUrl + "/auth/register"
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("createDoctor post: %w", err)
	}
	return res, nil
}

func newDoctor(email string) *api.DoctorRegistration {
	d := &api.DoctorRegistration{
		Email:          "dr.default@example.com",
		FirstName:      "Gregory",
		LastName:       "House",
		Specialization: api.Urologist,
		Role:           api.UserRoleDoctor,
	}
	if email != "" {
		d.Email = types.Email(email)
	}
	return d
}
