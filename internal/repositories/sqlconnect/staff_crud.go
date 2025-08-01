package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
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
		newStaff.Password, err = utils.HashPassword(newStaff.Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding staff into database")
		}

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
	setClauses := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	args = append(args, id)

	// Contoh: "first_name = ?, class = ?"
	query := fmt.Sprintf("UPDATE staff SET %s WHERE id = ?", strings.Join(setClauses, ", "))

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

func GetUserByUsername(username string) (*models.Staff, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}
	defer db.Close()

	user := &models.Staff{}
	err = db.QueryRow(`SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM staff WHERE username = ?`, username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Password, &user.InactiveStatus, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrorHandler(err, "user not found")
		}
		return nil, utils.ErrorHandler(err, "database error")
	}
	return user, nil
}

func UpdatePasswordInDb(userId int, currentPassword, newPassword string) (bool, error) {
	db, err := ConnectDb()
	if err != nil {
		return false, utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()

	var username string
	var userPassword string
	var userRole string

	err = db.QueryRow(`SELECT username, password, role FROM staff WHERE id = ?`, userId).Scan(&username, &userPassword, &userRole)
	if err != nil {
		return false, utils.ErrorHandler(err, "user not found")
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {

		return false, utils.ErrorHandler(err, "The password you entered does not match the current password on file")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {

		return false, utils.ErrorHandler(err, "internal error")
	}

	currentTime := time.Now().Format(time.RFC3339)

	_, err = db.Exec(`UPDATE staff SET password = ?, password_changed_at = ? WHERE id = ?`, hashedPassword, currentTime, userId)
	if err != nil {

		return false, utils.ErrorHandler(err, "failed to update the password")
	}

	// token, err := utils.SignToken(userId, username, userRole)
	// if err != nil {
	// 	utils.ErrorHandler(err, "Password update. Could not create token")
	// 	return
	// }

	return true, nil
}

func ForgotPasswordDbHandler(emailId string) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
	}
	defer db.Close()

	var staff models.Staff
	err = db.QueryRow(`SELECT id FROM staff WHERE email = ?`, emailId).Scan(&staff.ID)
	if err != nil {
		return utils.ErrorHandler(err, "User not found")
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}
	mins := time.Duration(duration)

	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}

	log.Println("tokenBytes:", tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	log.Println("token:", token)

	hashedToken := sha256.Sum256(tokenBytes)
	log.Println("hashedToken:", hashedToken)

	hashedTokenString := hex.EncodeToString(hashedToken[:])

	_, err = db.Exec(`UPDATE staff SET password_reset_token = ?, password_token_expires = ? WHERE id = ?`, hashedTokenString, expiry, staff.ID)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}

	// send the reset email
	resetURL := fmt.Sprintf("https://localhost:3000/staff/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Forgot your passsword? Reset your password using the following link: \n%s\nIf you didn't request a password reset, please ignore this email. This link is only valid for %d minutes.", resetURL, int(mins))

	m := mail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com") // sesuain
	m.SetHeader("To", emailId)
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)

	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}
	return nil
}
