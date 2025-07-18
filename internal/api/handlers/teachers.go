package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restapi/internal/models"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	teachers = make(map[int]models.Teacher)
 	mutex = &sync.Mutex{}
 	nextID = 1
)

// Initialize some dummy data
func init() {
	teachers [nextID] = models.Teacher{
		ID: nextID,
		FirstName: "Udin",
		LastName: "Ivana",
		Class: "SI1",
		Subject: "RPL",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID: nextID,
		FirstName: "Asep",
		LastName: "Ivana",
		Class: "SI2",
		Subject: "RPL",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID: nextID,
		FirstName: "Yanti",
		LastName: "Azmi",
		Class: "SI3",
		Subject: "RPL",
	}
	nextID++
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
	case http.MethodPost:
		addTeacherHandler(w, r)
	case http.MethodPut:
		fmt.Fprintf(w, "Hello PUT Method on Teachers Route")
		return
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Teachers Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Teachers Route")
		return
	}
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	fmt.Println(idStr)

	if idStr == "" {
		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		// teacherList := make([]Teacher, 0, len(teachers))
		// for _, teacher := range teachers {
		// 	if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
		// 		teacherList = append(teacherList, teacher)
		// 	}
		// }

		// --- MULAI PERUBAHAN DI SINI ---
    // 1. Kumpulkan semua kunci (ID) dari map
    keys := make([]int, 0, len(teachers))
    for k := range teachers {
        keys = append(keys, k)
    }

    // 2. Urutkan kunci (ID) dari kecil ke besar
    sort.Ints(keys)

    // 3. Buat teacherList dengan mengulang dari kunci yang sudah terurut
    teacherList := make([]models.Teacher, 0, len(teachers))
    for _, id := range keys {
        teacher := teachers[id] // Ambil guru berdasarkan ID yang terurut
        
        // Terapkan filter di sini
        if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
            teacherList = append(teacherList, teacher)
        }
    }
    // --- AKHIR PERUBAHAN ---

		response := struct {
			Status string `json:"status"`
			Count int `json:"count"`
			Data []models.Teacher `json:"data"`
		}{
			Status: "success",
			Count: len(teacherList),
			Data: teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)}

		// Handle Path parameter
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Println(err)
			return
		}

		teacher, exists := teachers[id]
		if !exists {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(teacher)
}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var newTeachers []models.Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		newTeacher.ID = nextID
		teachers[nextID] = newTeacher
		addedTeachers[i] = newTeacher
		nextID++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string `json:"status"`
		Count  int `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count: len(addedTeachers),
		Data: addedTeachers,
	}
	json.NewEncoder(w).Encode(response)

}