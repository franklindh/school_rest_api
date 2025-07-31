package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"strconv"
)

func GetStaffHandler(w http.ResponseWriter, r *http.Request) {

	var staff []models.Staff
	staff, err := sqlconnect.GetStaffDBHandler(staff, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string         `json:"status"`
		Count  int            `json:"count"`
		Data   []models.Staff `json:"data"`
	}{
		Status: "success",
		Count:  len(staff),
		Data:   staff,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOneStaffHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Id", http.StatusBadRequest)
		return
	}

	staff, err := sqlconnect.GetStaffByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staff)
}

func AddStaffHandler(w http.ResponseWriter, r *http.Request) {

	var newstaffs []models.Staff
	var rawstaffs []map[string]any

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawstaffs)

	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Staff{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, staff := range rawstaffs {
		for key := range staff {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}

	err = json.Unmarshal(body, &newstaffs)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	for _, staff := range newstaffs {
		err := CheckBlankFields(staff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	addedstaff, err := sqlconnect.AddStaffDBHandler(newstaffs)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string         `json:"status"`
		Count  int            `json:"count"`
		Data   []models.Staff `json:"data"`
	}{
		Status: "success",
		Count:  len(addedstaff),
		Data:   addedstaff,
	}
	json.NewEncoder(w).Encode(response)
}

func PatchStaffHandler(w http.ResponseWriter, r *http.Request) {

	var updates []map[string]any
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchStaff(updates)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func PatchOneStaffHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid staff Id", http.StatusBadRequest)
		return
	}

	var updates map[string]any
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	updatedstaff, err := sqlconnect.PatchOneStaff(id, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedstaff)
}

func DeleteOneStaffHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid staff Id", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneStaff(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "staff succesfully deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}
