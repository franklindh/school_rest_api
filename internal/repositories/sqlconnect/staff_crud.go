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

func GetStaffDBHandler(staffmodel []models.Staff, r *http.Request) ([]models.Staff, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM staff WHERE 1=1"
	var args []any

	query, args = utils.AddFilter(r, query, args)

	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	defer rows.Close()

	// StaffList := make([]models.Staff, 0)
	for rows.Next() {
		var staff models.Staff
		err := rows.Scan(&staff.ID, &staff.FirstName, &staff.LastName, &staff.Email, &staff.Username, &staff.UserCreatedAt, &staff.InactiveStatus, &staff.Role)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		staffmodel = append(staffmodel, staff)
	}
	return staffmodel, nil
}

func GetStaffByID(id int) (models.Staff, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Staff{}, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	var staff models.Staff
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM staff WHERE id = ?", id).Scan(&staff.ID, &staff.FirstName, &staff.LastName, &staff.Email, &staff.Username, &staff.UserCreatedAt, &staff.InactiveStatus, &staff.Role)
	if err == sql.ErrNoRows {
		return models.Staff{}, utils.ErrorHandler(err, "error retrieving data")
	} else if err != nil {
		return models.Staff{}, utils.ErrorHandler(err, "error retrieving data")
	}
	return staff, nil
}

func AddStaffDBHandler(newStaff []models.Staff) ([]models.Staff, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer db.Close()

	// stmt, err := db.Prepare("INSERT INTO Staffs(first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("staff", models.Staff{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error adding data")
	}
	defer stmt.Close()

	addedStaff := make([]models.Staff, len(newStaff))
	for i, newStaff := range newStaff {
		// res, err := stmt.Exec(newStaff.FirstName, newStaff.LastName, newStaff.Email, newStaff.Class, newStaff.Subject)
		values := utils.GetStructValues(newStaff)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding data")
		}
		newStaff.ID = int(lastID)
		addedStaff[i] = newStaff
	}
	return addedStaff, nil
}

func PatchStaff(updates []map[string]any) error {
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

		var staffFromDb models.Staff
		err = db.QueryRow("SELECT id, first_name, last_name, email, username FROM staff WHERE id = ?", id).Scan(&staffFromDb.ID, &staffFromDb.FirstName, &staffFromDb.LastName, &staffFromDb.Email, &staffFromDb.Username)
		if err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return utils.ErrorHandler(err, "Staff not found")
			}
			return utils.ErrorHandler(err, "error updating data")
		}

		staffVal := reflect.ValueOf(&staffFromDb).Elem()
		staffType := staffVal.Type()

		for key, value := range update {
			if key == "id" {
				continue // skip updating the id field
			}
			for i := 0; i < staffVal.NumField(); i++ {
				field := staffType.Field(i)
				if field.Tag.Get("json") == key+",omitempty" {
					fieldVal := staffVal.Field(i)
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
		_, err = tx.Exec("UPDATE staff SET first_name = ?, last_name = ?, email = ?, username = ? WHERE id = ?", staffFromDb.FirstName, staffFromDb.LastName, staffFromDb.Email, &staffFromDb.Username, staffFromDb.ID)

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

func PatchOneStaff(id int, updates map[string]any) (models.Staff, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Staff{}, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	var existingStaff models.Staff
	err = db.QueryRow("SELECT id, first_name, last_name, email, username FROM staff WHERE id = ?", id).Scan(&existingStaff.ID, &existingStaff.FirstName, &existingStaff.LastName, &existingStaff.Email, &existingStaff.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Staff{}, utils.ErrorHandler(err, "Staff not found")
		}
		return models.Staff{}, utils.ErrorHandler(err, "error updating data")
	}

	staffVal := reflect.ValueOf(&existingStaff).Elem()
	staffType := staffVal.Type()

	for key, value := range updates {
		for i := 0; i < staffVal.NumField(); i++ {
			field := staffType.Field(i)
			if field.Tag.Get("json") == key+",omitempty" {
				if staffVal.Field(i).CanSet() {
					fieldVal := staffVal.Field(i)
					fieldVal.Set(reflect.ValueOf(value).Convert(staffVal.Field(i).Type()))
				}
			}

		}
	}

	// _, err = db.Exec("UPDATE Staffs SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", existingStaff.FirstName, existingStaff.LastName, existingStaff.Email, existingStaff.Class, existingStaff.Subject, existingStaff.ID)
	// if err != nil {
	// 	// log.Println(err)
	// 	http.Error(w, "Error updating Staff", http.StatusInternalServerError)
	// 	return
	// }
	// Siapkan slice untuk menampung bagian SET dari query dan argumennya
	setClauses := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)

	// Bangun query secara dinamis dari map 'updates'
	for key, value := range updates {
		// Asumsi: key dari JSON sama dengan nama kolom di database
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	// Tambahkan ID ke akhir slice argumen untuk klausa WHERE
	args = append(args, id)

	// Gabungkan semua `setClauses` menjadi satu string, dipisahkan koma
	// Contoh: "first_name = ?, class = ?"
	query := fmt.Sprintf("UPDATE staff SET %s WHERE id = ?", strings.Join(setClauses, ", "))

	// Eksekusi query yang sudah dibangun secara dinamis
	_, err = db.Exec(query, args...)
	if err != nil {
		return models.Staff{}, utils.ErrorHandler(err, "error updating data")
	}
	return existingStaff, nil
}

func DeleteStaff(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting data")
	}

	stmt, err := tx.Prepare("DELETE FROM staff WHERE id = ?")
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

func DeleteOneStaff(id int) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM staff WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "error deleting data")
	}

	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Staff not found")
	}
	return nil
}
