package console

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RatRyan/dbapp/internal/employee"
	"github.com/RatRyan/dbapp/internal/serialize"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

var Running = true
var dirPath = "C:/Users/rratajczak/Documents/people/long"
var employeeMap = make(map[int]employee.Employee)

func deserialize(id string) employee.Employee {
	data, err := os.ReadFile(dirPath + " serialized/" + id + ".ser")
	if err != nil {
		log.Fatal("File doesn't exist: " + id + ".ser")
	}

	protoEmployee := serialize.Employee{}
	err = proto.Unmarshal(data, &protoEmployee)
	if err != nil {
		log.Fatal("protobuf error")
	}

	return employee.Employee{
		Id:        int(protoEmployee.GetId()),
		FirstName: protoEmployee.GetFirstName(),
		LastName:  protoEmployee.GetLastName(),
		HireDate:  int(protoEmployee.GetHireDate()),
	}
}

func findEmployeeByLastName(lastName string) (employee.Employee, error) {
	files, _ := os.ReadDir(dirPath)

	for _, file := range files {
		emp := deserialize(strings.TrimSuffix(file.Name(), path.Ext(file.Name())))
		if strings.EqualFold(emp.LastName, lastName) {
			return emp, nil
		}
	}

	return employee.Employee{}, errors.New("No employee found")
}

func findAllEmployeesByLastName(lastName string) []employee.Employee {
	empList := []employee.Employee{}
	files, _ := os.ReadDir(dirPath)

	for _, file := range files {
		emp := deserialize(strings.TrimSuffix(file.Name(), path.Ext(file.Name())))
		if strings.EqualFold(emp.LastName, lastName) {
			empList = append(empList, emp)
		}
	}

	return empList
}

func getAllEmployees() map[int]employee.Employee {
	employeeMap := make(map[int]employee.Employee)
	files, _ := os.ReadDir(dirPath)

	for i, file := range files {
		employeeMap[i] = deserialize(strings.TrimSuffix(file.Name(), path.Ext(file.Name())))
	}

	return employeeMap
}

var commands = []*cli.Command{
	{
		Name:      "path",
		UsageText: "path [command]",
		Usage:     "Prints out the current path",
		ArgsUsage: "Prints out the current path being used",
		Action: func(cCtx *cli.Context) error {
			fmt.Println("current path: " + dirPath)
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "set",
				Usage: "Sets the path to the people folder",
				Action: func(cCtx *cli.Context) error {
					dirPath = filepath.ToSlash(cCtx.Args().Get(0))

					fmt.Println("Path set to: '" + dirPath + "'")
					return nil
				},
			},
		},
	},
	{
		Name:        "employee",
		Usage:       "Provides tools for managing employees",
		Description: "used to add/update/delete employees from",
		Subcommands: []*cli.Command{
			{
				Name:        "add",
				Usage:       "Adds a new employee to the database",
				Description: "Adds a new employee to the database",
				Action: func(cCtx *cli.Context) error {
					args := cCtx.Args().Slice()

					id, _ := strconv.Atoi(args[0])
					hireDate, _ := strconv.Atoi(args[3])

					employee.WriteEmployee(id, args[1], args[2], hireDate, dirPath+"/")
					fmt.Println("Employee successfully added")
					return nil
				},
			},
			{
				Name:        "update",
				Usage:       "Updates an existing employee in the database",
				Description: "Updates an existing employee",
				Action: func(cCtx *cli.Context) error {
					args := cCtx.Args().Slice()

					id, _ := strconv.Atoi(args[0])
					hireDate, _ := strconv.Atoi(args[3])

					_, err := os.ReadFile(dirPath + "/" + strconv.Itoa(id) + ".txt")
					if err != nil {
						fmt.Println("File doesn't exist")
						return nil
					}

					employee.WriteEmployee(id, args[1], args[2], hireDate, dirPath+"/")
					fmt.Println("Employee " + strconv.Itoa(id) + " Successfully updated")
					return nil
				},
			},
			{
				Name:        "delete",
				Usage:       "Removes an employee from the database",
				Description: "deletes the employee with matching ID",
				Action: func(cCtx *cli.Context) error {
					err := os.Remove(dirPath + "/" + cCtx.Args().First() + ".txt")
					if err != nil {
						fmt.Println("file not found")
						return nil
					}

					fmt.Println("Employee " + cCtx.Args().First() + " successfully deleted")
					return nil
				},
			},
		},
	},
	{
		Name:        "print",
		Usage:       "Prints the raw data of every employee",
		Description: "prints out people details",
		Action: func(cCtx *cli.Context) error {
			start := time.Now()
			files, _ := os.ReadDir(dirPath)

			for _, file := range files {
				data, _ := os.ReadFile(path.Join(dirPath, file.Name()))
				fmt.Println(string(data))
			}

			fmt.Println("Elapsed time:", time.Since(start))
			return nil
		},
	},
	{
		Name:        "serialize",
		Usage:       "Serializes every file in the current path",
		Description: "serializes all of the files in the current path",
		Action: func(cCtx *cli.Context) error {

			dir, err := os.ReadDir(dirPath)
			if err != nil {
				fmt.Println("Directory doesn't exist")
				return nil
			}

			var wg sync.WaitGroup
			start := time.Now()

			for _, file := range dir {

				wg.Add(1)
				go func(entry os.DirEntry) {
					rawData, _ := os.ReadFile(filepath.Join(dirPath, entry.Name()))
					data := strings.Split(strings.TrimSpace(string(rawData)), ", ")

					id, _ := strconv.Atoi(data[0])
					hireDate, _ := strconv.Atoi(data[3])

					employee := &serialize.Employee{
						Id:        int64(id),
						FirstName: data[1],
						LastName:  data[2],
						HireDate:  int64(hireDate),
					}

					message, err := proto.Marshal(employee)
					if err != nil {
						fmt.Println("Protobuf error")
					}

					err = os.WriteFile(filepath.Join(dirPath+" serialized", data[0]+".ser"), message, 0777)
					if err != nil {
						fmt.Println("File failed to write")
					}

					wg.Done()
				}(file)
			}

			fmt.Println("Finished Serialization! Time Elapsed:", time.Since(start))
			return nil
		},
	},
	{
		Name:        "deserialize",
		Usage:       "Deserialize an employee by id",
		Description: "Deserialize an employee given an id",
		Action: func(cCtx *cli.Context) error {
			employee := deserialize(cCtx.Args().Get(0))

			fmt.Println(employee)
			return nil
		},
	},
	{
		Name:        "find",
		Usage:       "Finds the first employee by last name",
		Description: "prints out the first occurance of an employee with a matching last name",
		Action: func(cCtx *cli.Context) error {
			emp, err := findEmployeeByLastName(cCtx.Args().Get(0))
			if err != nil {
				fmt.Println("employee not found")
				return nil
			}

			fmt.Println(emp)
			return nil
		},
	},
	{
		Name:        "findAll",
		Usage:       "Finds all of the employees with a matching last name",
		Description: "prints out all of the employees with a matching last name",
		Action: func(cCtx *cli.Context) error {
			empList := findAllEmployeesByLastName(cCtx.Args().Get(0))

			fmt.Println(len(empList), "employees found")
			for _, emp := range empList {
				fmt.Println(emp)
			}

			return nil
		},
	},
	{
		Name:  "hashMap",
		Usage: "Sets up a map and prints out the employees",
		Action: func(cCtx *cli.Context) error {
			employeeMap = getAllEmployees()
			start := time.Now()

			for i := 0; i < len(employeeMap); i++ {
				fmt.Println(employeeMap[i])
			}

			fmt.Println("Elapsed time:", time.Since(start))
			return nil
		},
	},
	{
		Name:        "quit",
		Usage:       "Quits out of the application",
		Description: "uits out of the application",
		Action: func(cCtx *cli.Context) error {
			Running = false
			return nil
		},
	},
}

var App = &cli.App{
	Name:        "Database Application",
	Usage:       "A CRUD database tool",
	UsageText:   "[command]",
	Description: "Console Application for performing CRUD actions on a .txt based database",
	Commands:    commands,
}
