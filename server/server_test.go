package server

import (
	"API/db"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	dbTest, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}

	if err := dbTest.AutoMigrate(&db.Department{}, &db.Employee{}); err != nil {
		t.Fatalf("Ошибка миграции тестовой БД: %v", err)
	}

	dataBase = dbTest
}

func TestCreateDepartment(t *testing.T) {
	setupTestDB(t)

	payload := map[string]string{"name": "Department"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	CreateDepartment(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получили %d", resp.StatusCode)
	}

	var department db.Department
	json.NewDecoder(resp.Body).Decode(&department)

	if department.Name != "Department" {
		t.Fatalf("Ожидалось имя 'Department', получили '%s'", department.Name)
	}
}

func TestCreateEmployee(t *testing.T) {
	setupTestDB(t)

	department := db.Department{Name: "Department"}
	dataBase.Create(&department)

	payload := map[string]string{"full_name": "Name Surename", "position": "Position"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/departments/1/employees", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	CreateEmployee(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получили %d", resp.StatusCode)
	}
}

func TestGetDepartmentByID(t *testing.T) {
	setupTestDB(t)

	department := db.Department{Name: "Name"}
	dataBase.Create(&department)

	req := httptest.NewRequest(http.MethodGet, "/departments/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	GetDepartmentByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался 200, получили %d", resp.StatusCode)
	}

	var result db.Department
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Name != "Name" {
		t.Fatalf("Ожидалось 'Name', получили '%s'", result.Name)
	}
}

func TestUpdateDepartment(t *testing.T) {
	setupTestDB(t)

	department := db.Department{Name: "Backend"}
	dataBase.Create(&department)

	payload := map[string]string{"name": "Backend Core"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/departments/1", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	UpdateDepartment(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался 200, получили %d", resp.StatusCode)
	}

	var updated db.Department
	json.NewDecoder(resp.Body).Decode(&updated)

	if updated.Name != "Backend Core" {
		t.Fatalf("Имя не обновилось")
	}
}

func TestDeleteDepartmentCascade(t *testing.T) {
	setupTestDB(t)

	department := db.Department{Name: "Department"}
	dataBase.Create(&department)

	req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=cascade", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteDepartment(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Ожидался 204, получили %d", resp.StatusCode)
	}

	var count int64
	dataBase.Model(&db.Department{}).Count(&count)
	if count != 0 {
		t.Fatalf("Подразделение не удалилось")
	}
}

func TestDeleteDepartmentReassign(t *testing.T) {
	setupTestDB(t)

	from := db.Department{Name: "Department one"}
	to := db.Department{Name: "Department two"}
	dataBase.Create(&from)
	dataBase.Create(&to)

	emp := db.Employee{
		DepartmentID: from.ID,
		FullName:     "Name Surename",
		Position:     "Position",
	}
	dataBase.Create(&emp)

	req := httptest.NewRequest(http.MethodDelete, "/departments/1?mode=reassign&reassign_to_department_id=2", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	DeleteDepartment(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Ожидался 204, получили %d", resp.StatusCode)
	}

	var updatedEmp db.Employee
	dataBase.First(&updatedEmp, emp.ID)

	if updatedEmp.DepartmentID != to.ID {
		t.Fatalf("Сотрудник не был переназначен")
	}
}
