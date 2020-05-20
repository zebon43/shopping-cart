package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//Data Structure for storing Orders
type Order struct {
	OId          int     `json:"ID"`
	PID          int     `json:"PID"`
	PDesc 		string `json:"Description"`
	Quantity     float32 `json:"Quantity"`
	Discount     float32 `json:"Discount"`
	Coupon       string  `json:"Coupon"`
	OrderTotal   float32 `json:"OrderTotal"`
	OrderSavings float32 `json:"OrderSavings"`
}

//Data Structure for storing Products
type Product struct {
	PID int
	Name string
	Price float32
}

//Data Structure for storing summary
type Summary struct {
	Total float32 `json:"Total"`
	Savings float32 `json:"Savings"`
}

//Data Structure for storing users
type User struct {
	UserID string `json:"UserID"`
	Password string `json:"Password"`
	Session bool
}

//Data Structure for storing cards
type Card struct {
	CardNo int `json:"CardNo"`
	ExpiryM int `json:"ExpiryM"`
	ExpiryY int `json:"ExpiryY"`
	CVV int `json:"CVV"`
}

//Global Variables for storing the required data
var orderCount int
var Orders []Order
var summary Summary
var UsedCodes = []string{"Dummy"}
var validCodes = []string{"O30"}
var users = []User{{"admin","admin", false}}
var cards = []Card{{1234567812345678,06,2020, 987}}
var Products = []Product{
	{0,"Apples",5.00},
	{1,"Bananas",10.00},
	{2,"Pears",15.00},
	{3,"Oranges",20.00},
}

//Display the login page to the user
func loginPage(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | Login Page is requested.")

	//Parse the template files for the UI
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {log.Println(" Unable to parse template files. |", err)}

	err = tmpl.Execute(w, "A")
	if err != nil {log.Println(" Unable to execute template files. |", err)}
}

//Display the shopping cart to the user
func cartPage(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | Cart Page is requested.")

	//Parse the template files for the UI
	tmpl, err := template.ParseFiles("templates/cart.html")
	if err != nil {log.Println(" Unable to parse template files. |", err)}

	err = tmpl.Execute(w, "A")
	if err != nil {log.Println(" Unable to execute template files. |", err)}
}

//Display the payment page to the user
func paymentPage(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | Payment Page is requested.")

	//Parse the template files for the UI
	tmpl, err := template.ParseFiles("templates/payment.html")
	if err != nil {log.Println(" Unable to parse template files. |", err)}

	err = tmpl.Execute(w, "A")
	if err != nil {log.Println(" Unable to execute template files. |", err)}
}

func userAuth(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | User Authentication is requested.")

	var user User
	//Get the JSON from the request Body
	tempUser, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(tempUser, &user)
	if err != nil {log.Println(" Unable to marshal incoming json. |", err)}

	if user.UserID != "" || user.Password != "" {
		//Validate Credentials
		for _ , each := range users {
			if each.UserID == user.UserID && each.Password == user.Password {
				http.Redirect(w, r, "/cart", 301)
				break
			} else {
				w.WriteHeader(401)
				if _ ,err := w.Write([]byte("Incorrect Credentials.")); err != nil {
					panic(err)
				}
			}
		}
	}
}

//Return all the orders for the user
func returnOrders(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | All orders requested.")
	if err := json.NewEncoder(w).Encode(Orders); err != nil {
		panic(err)
	}
}

//Create the orders for the user
func createOrder(w http.ResponseWriter, r *http.Request) {
	orderCount++
	orderExists := false //flag to check is the order is present
	var order Order //Store the json data

	log.Println( r.Method, "| New Order requested.",)

	//Get the JSON from the request Body
	tempOrder, _ := ioutil.ReadAll(r.Body)

	//Convert the json to struct
	err := json.Unmarshal(tempOrder, &order)
	if err != nil {log.Println(" Unable to marshal incoming json. |", err)}

	//Create the new order data
	for i , each := range Orders {
		if each.PID == order.PID {
			Orders[i].Quantity = order.Quantity
			Orders[i].OrderTotal = order.OrderTotal
			Orders[i].Coupon = order.Coupon
			orderExists = true
			i++
		}
	}

	if orderExists == false {
		order.OId = orderCount
		Orders = append(Orders, order)
	}

	//Calculate the total and the discount for each order
	calculate()

	//Return the order in json format
	if err := json.NewEncoder(w).Encode(Orders); err != nil {
		panic(err)
	}
}

//Calculate the total and the discount
func calculate() string {
	applyCoupon := false
	status := ""
	for i , each := range Orders {

		Orders[i].Discount = 0
		if Orders[i].PID != 3{Orders[i].Coupon = ""}

		//Discount calculation for apples
		if each.PID == 0 && each.Quantity >= 7 {
			Orders[i].Discount = 10
			status = "Discount Applied to Apples."
		} else if each.PID == 1 && each.Quantity >= 2 { //Discount calculation for Pears and Bananas
			for j , each1 := range Orders {
				if each1.PID == 2 && each1.Quantity >= 4 {
					Orders[i].Discount = 30
					Orders[j].Discount = 30
					status = "Discount Applied to Pear & Banana set."
				}
			}
		} else{ //Discount calculation for Oranges using codes
			for _ , k := range validCodes {
				if each.PID == 3 && k == each.Coupon {
					for _ , l := range UsedCodes{
						if each.PID == 3 && l == each.Coupon {
							applyCoupon = false
							status = "PromoCode is already used."
							break
						} else {
							applyCoupon = true
							Orders[i].Coupon = each.Coupon
						}
					}
				} else {
					applyCoupon = false
				}
			}
			if applyCoupon == true {
				Orders[i].Discount = 30
				status = "PromoCode applied."
			}
		}

		Orders[i].OrderTotal = each.Quantity * Products[each.PID].Price
		Orders[i].OrderSavings = Orders[i].OrderTotal * Orders[i].Discount / 100
	}
	if status != "" {log.Println( "Calculate| ",status)}
	return status
}

//calculate the total sum for all orders
func sum(w http.ResponseWriter, r *http.Request) {

	summary.Total, summary.Savings = 0, 0
	for _ , each := range Orders{
		summary.Total = summary.Total + each.OrderTotal - each.OrderSavings
		summary.Savings = summary.Savings + each.OrderSavings
	}

	//Return the summary in json format
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		panic(err)
	}
}

//Delete orders from cart
func deleteOrder(w http.ResponseWriter, r *http.Request) {

	var order Order
	log.Println( r.Method, "| Delete Order requested.",)

	//Get the JSON from the request Body
	tempOrder, _ := ioutil.ReadAll(r.Body)

	//Convert the json to struct
	err := json.Unmarshal(tempOrder, &order)
	if err != nil {log.Println(" Unable to marshal incoming json. |", err)}

	//Find the order to be delete
	for i, each := range Orders{
		if each.PID == order.PID {
			//Delete the request order
			copy(Orders[i:], Orders[i+1:])
			//Orders[len(Orders)-1] = ""
			Orders = Orders[:len(Orders)-1]
			break
		}
		i++
	}

	//Return the update order list in json format
	if err := json.NewEncoder(w).Encode(Orders); err != nil {
		panic(err)
	}
}

//Checkout orders from cart for payment
func checkoutOrder(w http.ResponseWriter, r *http.Request){
	log.Println( r.Method, " | Checkout is requested.")
	calculate()
	http.Redirect(w, r, "/payment",200)
}

//Payment for the orders
func processPayment(w http.ResponseWriter, r *http.Request) {

	var card Card
	log.Println( r.Method, "| Checkout for Order requested.",)

	//Get the JSON from the request Body
	tempCard, _ := ioutil.ReadAll(r.Body)

	//Convert the json to struct
	err := json.Unmarshal(tempCard, &card)
	if err != nil {log.Println(" Unable to marshal incoming json. |", err)}

	//Validate the card
	for _, each := range cards{
		if each.CardNo == card.CardNo && each.ExpiryM == card.ExpiryM && each.ExpiryY == card.ExpiryY && each.CVV == card.CVV {
			if card.ExpiryY >= time.Now().Year() {
				if time.Month(card.ExpiryM) >= time.Now().Month() {
					//Add any promocode applied to the used codes data structure
					for _ , each:= range Orders {
						if each.Coupon != "" {UsedCodes = append(UsedCodes, each.Coupon)}
					}
					Orders = nil
					w.WriteHeader(200)
					if _, err := w.Write([]byte("Payment Successful.")); err != nil {
						panic(err)
					}
					break
				}
			} else {
				w.WriteHeader(401)
				if _, err := w.Write([]byte("Card Expired.")); err != nil {
					panic(err)
				}
			}

		} else{
			w.WriteHeader(401)
			if _, err := w.Write([]byte("Invalid Card details.")); err != nil {
				panic(err)
			}
		}
	}
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	//set the path for css and image folders
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/",http.FileServer(http.Dir("templates/css/"))))
	router.PathPrefix("/img/").Handler(http.StripPrefix("/img/",http.FileServer(http.Dir("templates/img/"))))

	//routers for the application
	router.HandleFunc("/", loginPage)
	router.HandleFunc("/userAuth", userAuth)
	router.HandleFunc("/cart", cartPage)
	router.HandleFunc("/payment", paymentPage)
	router.HandleFunc("/orders", returnOrders)
	router.HandleFunc("/getsum", sum)
	router.HandleFunc("/createOrder", createOrder).Methods("POST")
	router.HandleFunc("/deleteOrder", deleteOrder).Methods("DELETE")
	router.HandleFunc("/checkout", checkoutOrder).Methods("POST")
	router.HandleFunc("/processpayment", processPayment).Methods("POST")

	//Log in case there is an error while the service is running
	log.Fatal(http.ListenAndServe(":8000", router))
}

func main() {
	log.Println("Shopping Cart Application Started.")
	handleRequests()
	log.Println("Shopping Cart Application Stopped.")
}