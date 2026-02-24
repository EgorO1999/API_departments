package server

import (
	"API/db"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var dataBase *gorm.DB

func Init(dbInstance *gorm.DB) {
	dataBase = dbInstance
}

func Run() {
	http.ListenAndServe(":8080", nil)
}

func InitRoutes() {
	router := mux.NewRouter()

	router.HandleFunc("/departments", CreateDepartment).Methods("POST")
	router.HandleFunc("/departments/{id}", GetDepartmentByID).Methods("GET")
	router.HandleFunc("/departments/{id}", UpdateDepartment).Methods("PATCH")
	router.HandleFunc("/departments/{id}", DeleteDepartment).Methods("DELETE")

	router.HandleFunc("/departments/{id}/employees", CreateEmployee).Methods("POST")

	http.Handle("/", router)
}

func CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		ParentID *int   `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		http.Error(w, "Название не может быть пустым", http.StatusBadRequest)
		return
	}

	if input.ParentID != nil {
		var parent db.Department
		if err := dataBase.First(&parent, *input.ParentID).Error; err != nil {
			http.Error(w, "Родительское подразделение не найдено", http.StatusNotFound)
			return
		}
	}

	department := db.Department{
		Name:     input.Name,
		ParentID: input.ParentID,
	}

	if err := dataBase.Create(&department).Error; err != nil {
		http.Error(w, "Подразделение с таким именем уже существует", http.StatusConflict)
		return
	}

	json.NewEncoder(w).Encode(department)
}

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	departmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный ID подразделения", http.StatusBadRequest)
		return
	}

	var department db.Department
	if err := dataBase.First(&department, departmentID).Error; err != nil {
		http.Error(w, "Подразделение не найдено", http.StatusNotFound)
		return
	}

	var input struct {
		FullName string  `json:"full_name"`
		Position string  `json:"position"`
		HiredAt  *string `json:"hired_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	input.FullName = strings.TrimSpace(input.FullName)
	input.Position = strings.TrimSpace(input.Position)

	if input.FullName == "" || input.Position == "" {
		http.Error(w, "full_name и position обязательны", http.StatusBadRequest)
		return
	}

	employee := db.Employee{
		DepartmentID: int(departmentID),
		FullName:     input.FullName,
		Position:     input.Position,
	}

	dataBase.Create(&employee)
	json.NewEncoder(w).Encode(employee)
}

func GetDepartmentByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	departmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	var department db.Department
	err = dataBase.Preload("Employees").Preload("Children").First(&department, departmentID).Error

	if err != nil {
		http.Error(w, "Подразделение не найдено", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(department)
}

func UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	departmentID, _ := strconv.Atoi(id)

	var department db.Department
	if err := dataBase.First(&department, departmentID).Error; err != nil {
		http.Error(w, "Подразделение не найдено", http.StatusNotFound)
		return
	}

	var input struct {
		Name     *string `json:"name"`
		ParentID *int    `json:"parent_id"`
	}

	json.NewDecoder(r.Body).Decode(&input)

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			http.Error(w, "Название не может быть пустым", http.StatusBadRequest)
			return
		}
		department.Name = name
	}

	if input.ParentID != nil {
		if *input.ParentID == department.ID {
			http.Error(w, "Нельзя сделать подразделение родителем самого себя", http.StatusConflict)
			return
		}
		department.ParentID = input.ParentID
	}

	dataBase.Save(&department)
	json.NewEncoder(w).Encode(department)
}

func DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	mode := r.URL.Query().Get("mode")

	departmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	switch mode {
	case "cascade":
		dataBase.Delete(&db.Department{}, departmentID)
		w.WriteHeader(http.StatusNoContent)

	case "reassign":
		tempStr := r.URL.Query().Get("reassign_to_department_id")
		if tempStr == "" {
			http.Error(w, "reassign_to_department_id обязателен", http.StatusBadRequest)
			return
		}

		reassignID, err := strconv.Atoi(tempStr)
		if err != nil {
			http.Error(w, "Неверный ID целевого подразделения", http.StatusBadRequest)
			return
		}

		tx := dataBase.Begin()

		tx.Model(&db.Employee{}).Where("department_id = ?", departmentID).Update("department_id", reassignID)

		tx.Delete(&db.Department{}, departmentID)

		tx.Commit()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Некорректный mode", http.StatusBadRequest)
	}
}
