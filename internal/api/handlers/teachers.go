package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"strconv"
	"strings"
)

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
	case http.MethodPost:
		addTeacherHandler(w, r)
	case http.MethodPut:
		
		
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Teachers Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Teachers Route")
		return
	}
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool {
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return  validFields[field]
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	fmt.Println(idStr)

	if idStr == "" {
		query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
		var args []any

		query, args = addFilter(r, query, args)

		query = addSorting(r, query)

		rows, err := db.Query(query, args...)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Database Quer Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

    teacherList := make([]models.Teacher, 0)
		for rows.Next() {
			var teacher models.Teacher
			err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
			if err != nil {
				http.Error(w, "Error scanning database result", http.StatusInternalServerError)
				return 
			}
			teacherList = append(teacherList, teacher)
		}
    
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

		var teacher models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return 
		} else if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(teacher)
}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilter(r *http.Request, query string, args []any) (string, []any) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}

	for params, dbField := range params {
		value := r.URL.Query().Get(params)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Error in preparing SQL query", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Error getting last insert ID", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
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

// func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
// 	db, err := sqlconnect.ConnectDb()
// 	if err != nil {
// 		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
// 	}
// 	defer db.Close()

// 	var newTeachers []models.Teacher
// 	err = json.NewDecoder(r.Body).Decode(&newTeachers)
// 	if err != nil {
// 		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
// 		return
// 	}

// 	// 2. Mulai transaksi database.
// 	tx, err := db.Begin()
// 	if err != nil {
// 		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
// 		return
// 	}
// 	// 3. Defer Rollback sebagai jaring pengaman.
// 	//    Jika ada error di tengah jalan, semua perubahan akan dibatalkan.
// 	defer tx.Rollback()

// 	// 4. Prepare statement dari transaksi (tx), bukan dari db.
// 	stmt, err := tx.Prepare("INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")
// 	if err != nil {
// 		http.Error(w, "Error in preparing SQL query", http.StatusInternalServerError)
// 		return
// 	}
// 	defer stmt.Close()

// 	addedTeachers := make([]models.Teacher, len(newTeachers))
// 	for i, newTeacher := range newTeachers {
// 		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
// 		if err != nil {
// 			// Jika error, defer tx.Rollback() akan otomatis membatalkan semua insert sebelumnya.
// 			http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
// 			return
// 		}
// 		lastID, err := res.LastInsertId()
// 		if err != nil {
// 			http.Error(w, "Error getting last insert ID", http.StatusInternalServerError)
// 			return
// 		}
// 		newTeacher.ID = int(lastID)
// 		addedTeachers[i] = newTeacher
// 	}

// 	// 5. Jika semua loop berhasil, simpan semua perubahan secara permanen.
// 	if err := tx.Commit(); err != nil {
// 		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
// 		return
// 	}

// 	// Siapkan dan kirim response sukses.
// 	response := struct {
// 		Status string             `json:"status"`
// 		Count  int                `json:"count"`
// 		Data   []models.Teacher   `json:"data"`
// 	}{
// 		Status: "success",
// 		Count:  len(addedTeachers),
// 		Data:   addedTeachers,
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(response)
// }
func updateTeacherHandler(){
	
}