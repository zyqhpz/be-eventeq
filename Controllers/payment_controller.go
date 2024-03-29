package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
)

type Item struct {
	ItemID 		primitive.ObjectID 	`bson:"id"`
	Name		string				`bson:"name"`
	Price		float64 			`bson:"price"`
	Quantity 	int32 				`bson:"quantity"`
}

type Booking struct {
	ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
	UserID 		primitive.ObjectID 	`bson:"user_id"`
	OwnerID		primitive.ObjectID 	`bson:"owner_id"`
	Items 		[]Item 				`bson:"items"`
	StartDate 	string 				`bson:"start_date"`
	EndDate 	string 				`bson:"end_date"`
	SubTotal 	float64 			`bson:"sub_total"`
	ServiceFee 	float64 			`bson:"service_fee"`
	GrandTotal 	float64 			`bson:"grand_total"`
	Status 		int32 				`bson:"status"`
	BillCode	string 				`bson:"bill_code"`
	CreatedAt 	time.Time 			`bson:"created_at"`
	UpdatedAt 	time.Time 			`bson:"updated_at"`
}

func CreatePaymentBillCode(booking *Booking) (string, error) {

	redirectUrl := "https://fe-eventeq.vercel.app/payment/redirect"
	// redirectUrl := "localhost:5173/payment/redirect"
	callbackUrl := "https://api.eventeq.xyz/api/payment/callback"

	// convert booking.GrandTotal to cents
	amount := booking.GrandTotal * 100

	// get user first and last name based on booking.UserID
	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	usersCollection := ConnectDBUsers(client)
	ctx := context.Background()

	filter := bson.M{"_id": booking.UserID}

	var user model.User
	err = usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		log.Print("Error getting user:")
		log.Fatal(err)
	}

	// convert booking.ID from primitive.ObjectID to string
	bId := booking.ID.Hex()

	data := url.Values{
		"userSecretKey":          	{"kl09dapk-hy3u-qo56-xor8-skza9c3haaul"},
		"categoryCode":           	{"8gmtq198"},
		"billName":               	{"EventEQ Payment"},
		"billDescription":        	{"Booking ID: " + bId + " for " + user.FirstName + " " + user.LastName},
		"billPriceSetting":       	{"1"},
		"billPayorInfo":          	{"1"},
		"billAmount":             	{fmt.Sprintf("%f", amount)},
		"billReturnUrl":          	{redirectUrl},
		"billCallbackUrl":        	{callbackUrl},
		"billExternalReferenceNo": 	{bId},
		"billTo":                 	{user.FirstName + " " + user.LastName},
		"billEmail":             	{user.Email},
		"billPhone":              	{user.NoPhone},
		"billSplitPayment":       	{"0"},
		"billSplitPaymentArgs":   	{""},
		"billPaymentChannel":     	{"0"},
		"billContentEmail":       	{"Thank you for purchasing our product!"},
		"billChargeToCustomer":   	{"1"},
		"billExpiryDate":         	{""},
		"billExpiryDays":         	{""},
	}

	netClient := http.Client{Timeout: time.Second * 10}
	url := "https://dev.toyyibpay.com/index.php/api/createBill"

	resp, err := netClient.PostForm(url, data)
	if err != nil {
		fmt.Println("Error making request:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status:", resp.Status)
		return "", err
	}

	var result string
	_, err = fmt.Fscan(resp.Body, &result)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	// remove [] from result
	result = result[1 : len(result)-1]

	// get BlllCode from result
	var response struct {
		BillCode string `json:"BillCode"`
	}

	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		log.Fatal(err)
	}

	return response.BillCode, nil
}

func HandleRedirectUrl(c *fiber.Ctx) error {

	// rediirect URL Parameter ?status_id=1&billcode=bcweidjq&order_id=AFR341DFI&msg=ok&transaction_id=TP2308153866893011

	statusId := c.Query("status_id") // 1=success, 2=pending, 3=fail
	billCode := c.Query("billcode") // billcode
	orderId := c.Query("order_id") // order_id
	msg := c.Query("msg")

	log.Print(statusId)
	log.Print(billCode)
	log.Print(orderId)
	log.Print(msg)

	return c.JSON(fiber.Map{
		"status": "success",
	})
}

func HandleCallbackUrl(c *fiber.Ctx) error {
	/*
	Please note that Callback URL cannot be received in localhost. Callback URL will be supplied with the following datas in POST format:-

	refno : Payment reference no

	status : Payment status. 1= success, 2=pending, 3=fail

	reason : Reason for the status received

	billcode : Your billcode / permanent link

	order_id : Your external payment reference no, if specified

	amount : Payment amount received

	transaction_time : Datetime of the transaction status received.
	*/

	// get data from request body
	type body struct {
		RefNo 			string 	`json:"refno"`
		Status 			string 	`json:"status"`
		Reason 			string 	`json:"reason"`
		BillCode 		string 	`json:"billcode"`
		OrderId 		string 	`json:"order_id"`
		Amount 			string 	`json:"amount"`
		TransactionTime string 	`json:"transaction_time"`
	}

	var req body
	err := c.BodyParser(&req)
	if err != nil {
		log.Println("Error parsing JSON request body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// convert string to int
	status, err := strconv.Atoi(req.Status)
	err = UpdatePaymentStatus(req.BillCode, status)

	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(fiber.Map{
		"status": "success",
	})
}

func UpdatePaymentStatus(billCode string, status int) error { 

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	bookingsCollection := ConnectDBBookings(client)
	ctx := context.Background()

	filter := bson.M{"bill_code": billCode}

	if status == 1 {
		status = 0
		GenerateEmail(billCode)
	} else {
		status = -1
	}

	log.Println(billCode + " " + strconv.Itoa(status))

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err = bookingsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func PaymentStatusChecker(billCode string) error { 

	someData := map[string]string{
		"billCode": billCode,
	}

	data := bytes.NewBuffer([]byte(``))
	for key, value := range someData {
		data.WriteString(key)
		data.WriteString("=")
		data.WriteString(value)
		data.WriteString("&")
	}

	client := &http.Client{}
	url := "https://dev.toyyibpay.com/index.php/api/getBillTransactions"

	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	result := new(bytes.Buffer)
	_, err = result.ReadFrom(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return err
	}

	var response []map[string]interface{}

	if err := json.Unmarshal([]byte(result.String()), &response); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return err
	}

	// Extract 'billpaymentStatus' from the first element (assuming there's only one)
	if len(response) > 0 {
		billpaymentStatus, ok := response[0]["billpaymentStatus"].(string)
		if ok {
			if (billpaymentStatus == "1") {
				err = UpdatePaymentStatusByChecker(billCode)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			fmt.Println("billpaymentStatus not found or not a string")
		}
	}

	return nil
}

func UpdatePaymentStatusByChecker(billCode string) error { 

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	bookingsCollection := ConnectDBBookings(client)
	ctx := context.Background()

	filter := bson.M{"bill_code": billCode}

	update := bson.M{
		"$set": bson.M{
			"status": 0,
		},
	}

	_, err = bookingsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}