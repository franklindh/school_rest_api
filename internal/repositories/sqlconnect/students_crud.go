package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strings"
)

func GetStudentsDBHandler(students []models.Student, r *http.Request, limit, page int) ([]models.Student, int, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class FROM students WHERE 1=1`
	var args []any

	query, args = utils.AddFilter(r, query, args)

	// add pagination
	offset := (page - 1) * limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		return nil, 0, utils.ErrorHandler(err, "error retrieving data")
	}
	defer rows.Close()

	// StudentList := make([]models.Student, 0)
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, 0, utils.ErrorHandler(err, "error retrieving data")
		}
		students = append(students, student)
	}

	// get total count of students
	var totalStudents int
	err = db.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&totalStudents)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalStudents = 0
	}
	return students, totalStudents, nil
}

func GetStudentByID(id int) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	var student models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	return student, nil
}

func AddStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO Students(first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		// res, err := stmt.Exec(newStudent.FirstName, newStudent.LastName, newStudent.Email, newStudent.Class, newStudent.Subject)
		values := utils.GetStructValues(newStudent)
		res, err := stmt.Exec(values...)
		if err != nil {
			if strings.Contains(err.Error(), "a foreign key constraint fails (`school`.`students`, CONSTRAINT `students_ibfk_1` FOREIGN KEY (`class`) REFERENCES `teachers` (`class`))") {
				return nil, utils.ErrorHandler(err, "class/class teacher does not exist")
			}
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		newStudent.ID = int(lastID)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}

func UpdateStudent(id int, updatedStudent models.Student) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "error updating data")
		}
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}

	updatedStudent.ID = existingStudent.ID
	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", updatedStudent.FirstName, updatedStudent.LastName, updatedStudent.Email, updatedStudent.Class, updatedStudent.ID)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	return updatedStudent, nil
}

func PatchStudents(updates []map[string]any) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}

	for _, update := range updates {
		idFloat, ok := update["id"].(float64)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "invalid Id")
		}

		id := int(idFloat)

		var studentFromDb models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&studentFromDb.ID, &studentFromDb.FirstName, &studentFromDb.LastName, &studentFromDb.Email, &studentFromDb.Class)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Student not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		studentVal := reflect.ValueOf(&studentFromDb).Elem()
		studentType := studentVal.Type()

		for key, value := range update {
			if key == "id" {
				continue // skip updating the id field
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == key+",omitempty" {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(value)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Cannot conver %v to %v", val.Type(), fieldVal.Type())
							return utils.ErrorHandler(err, "error updating data")
						}
					}
					break
				}
			}
		}
		_, err = tx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", studentFromDb.FirstName, studentFromDb.LastName, studentFromDb.Email, studentFromDb.Class, studentFromDb.ID)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "error updating data")
		}
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "error updating data")
	}
	return nil
}

func PatchOneStudent(id int, updates map[string]any) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Student{}, utils.ErrorHandler(err, "Student not found")
		}
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}

	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()

	for key, value := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			if field.Tag.Get("json") == key+",omitempty" {
				if studentVal.Field(i).CanSet() {
					fieldVal := studentVal.Field(i)
					fieldVal.Set(reflect.ValueOf(value).Convert(studentVal.Field(i).Type()))
				}
			}

		}
	}

	// _, err = db.Exec("UPDATE Students SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", existingStudent.FirstName, existingStudent.LastName, existingStudent.Email, existingStudent.Class, existingStudent.Subject, existingStudent.ID)
	// if err != nil {
	// 	// log.Println(err)
	// 	http.Error(w, "Error updating Student", http.StatusInternalServerError)
	// 	return
	// }
	setClauses := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE students SET %s WHERE id = ?", strings.Join(setClauses, ", "))

	_, err = db.Exec(query, args...)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "error updating data")
	}
	return existingStudent, nil
}

func DeleteStudents(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer stmt.Close()

	deletedIds := []int{}

	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "error deleting data")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "error deleting data")
		}

		// if rowsAffected > 0
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID %d not found", id))
		}
	}

	// commit
	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	if len(deletedIds) < 1 {
		return nil, utils.ErrorHandler(err, "IDs do not exist")
	}
	return deletedIds, nil
}

func DeleteOneStudent(id int) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Student not found")
	}
	return nil
}

func GetStudentsByTeacherFromDB(teacherId string, students []models.Student) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	query := `SELECT id, first_name, last_name, email, class FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)`
	rows, err := db.Query(query, teacherId)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {

			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		students = append(students, student)
	}
	err = rows.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	return students, nil
}

func GetStudentCountByTeacherIdFromDB(teacherId string) (int, error) {
	db, err := ConnectDb()
	if err != nil {
		return 0, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	query := `SELECT COUNT(*) FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)`

	var studentCount int

	err = db.QueryRow(query, teacherId).Scan(&studentCount)
	if err != nil {
		return 0, utils.ErrorHandler(err, "error retrieving data")
	}
	return studentCount, nil
}
