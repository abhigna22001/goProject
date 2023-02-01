package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, dbname, password, host, port string) {

	connStr := "user=postgres dbname=postgres password=admin host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	a.DB = db
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	fmt.Println("Server is running at port 8080")
	cors := cors.New(cors.Options{AllowedOrigins: []string{"*"}, AllowedMethods: []string{http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete}, AllowedHeaders: []string{"*"}, AllowCredentials: false})
	fmt.Println("Server at 8080")
	http.Handle("/", a.Router)
	handler := cors.Handler(a.Router)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal("There's an error with the server", err)
	}
	fmt.Println("Server at 8080")
}

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_NAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_HOST"),
		os.Getenv("APP_DB_PORT"))
	a.Run(":8080")

}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/employees", a.getEmployees).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/employees/{empId}", a.getEmployeeById).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/employees", a.createEmployee).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/employees/{empId}", a.deleteEmployee).Methods("DELETE", "OPTIONS")
	a.Router.HandleFunc("/employees/{empId}", a.updateEmployee).Methods("PUT", "OPTIONS")
}

type employeeInfo struct {
	EmpId     int    `json:"emp_id"`
	EmpName   string `json:"emp_name"`
	EmpRole   string `json:"emp_role"`
	EmpSalary int    `json:"emp_salary"`
}

func (a *App) getEmployees(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	rows, err := a.DB.Query("SELECT * FROM employee")
	if err != nil {
		panic(err)
	}
	var employeeList []employeeInfo
	for rows.Next() {
		var empId int
		var empName string
		var empRole string
		var empSalary int
		err = rows.Scan(&empId, &empName, &empRole, &empSalary)
		if err != nil {
			panic(err)
		}
		employeeList = append(employeeList, employeeInfo{EmpId: empId, EmpName: empName, EmpRole: empRole, EmpSalary: empSalary})
	}
	json.NewEncoder(writer).Encode(employeeList)
}

func (a *App) getEmployeeById(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	employeeID := params["empId"]
	result, err := a.DB.Query("SELECT * FROM employee where empid=$1", employeeID)
	if err != nil {
		panic(err)
	}
	var employeeDetails employeeInfo
	for result.Next() {
		err = result.Scan(&employeeDetails.EmpId, &employeeDetails.EmpName, &employeeDetails.EmpRole, &employeeDetails.EmpSalary)
		if err != nil {
			panic(err)
		}
	}
	json.NewEncoder(writer).Encode(employeeDetails)
}

func (a *App) createEmployee(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	empName := keyVal["emp_name"]
	empRole := keyVal["emp_role"]
	empSalary := keyVal["emp_salary"]
	_, err = a.DB.Exec(`INSERT into employee("empname","emprole","empsalary") VALUES ($1,$2,$3)`, empName, empRole, empSalary)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(writer, "New user was created")
}

func (a *App) deleteEmployee(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	_, err := a.DB.Exec("DELETE FROM employee WHERE empid = $1 ", params["empId"])
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(writer, "User with ID = %s was deleted",
		params["empId"])
}

func (a *App) updateEmployee(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	empID := params["empId"]
	empName := keyVal["emp_name"]
	empRole := keyVal["emp_role"]
	empSalary := keyVal["emp_salary"]
	_, err = a.DB.Exec(`update employee set "empname"=$1 ,"emprole"=$2,"empsalary"=$3 where "empid"=$4`, empName, empRole, empSalary, empID)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(writer, "Employee details updated successfully!")
}
