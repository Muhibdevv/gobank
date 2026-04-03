package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Transaction struct {
	TransactionTime time.Time `json:"time"`
	TransactionType string    `json:"type"`
	Amount          float64   `json:"amount"`
}
type Accounts struct {
	HolderName         string        `json:"holdername"`
	AccountNumber      string        `json:"accountnumber"`
	AccountPIN         string        `json:"pin"`
	Balance            float64       `json:"balance"`
	TransactionHistory []Transaction `json:"history"`
}

var fileName = "accounts.json"

func menu() {
	var userInput int
	for {
		fmt.Println("=== Welcome to GOBank! ===")
		fmt.Println("1. Create Account")
		fmt.Println("2. Deposit Money")
		fmt.Println("3. Withdraw Money")
		fmt.Println("4. Check Balance")
		fmt.Println("5. Transfer Money")
		fmt.Println("6. Change PIN")
		fmt.Println("7. Transaction History")
		fmt.Println("8. Exit")
		fmt.Scanln(&userInput)

		switch userInput {
		case 1:
			if err := CreateAccount(); err != nil {
				fmt.Println("Error:", err)
			}
		case 2, 3, 4, 5, 6, 7:
			Account, err := Authenticate()
			if err != nil {
				fmt.Println("Error", err)
				continue
			}
			fmt.Println("Authentication Successful.")
			switch userInput {
			case 2:
				if err := Deposit(Account); err != nil {
					fmt.Println("Error:", err)
				}
			case 3:
				if err := Withdraw(Account); err != nil {
					fmt.Println("Error:", err)
				}
			case 4:
				FetchBalance(Account)
			case 5:
				if err := Transfer(Account); err != nil {
					fmt.Println("Error:", err)
				}
			case 6:
				if err := ChangePin(Account); err != nil {
					fmt.Println("Error:", err)
				}
			case 7:
				FetchHistory(Account)
			}

		case 8:
			return
		}
	}
}
func CreateAccount() error {
	var name string
	var pin string
	var initBalance float64

	fmt.Println("Please enter your name without space:")
	fmt.Scanln(&name)
	fmt.Println("Please add your initial deposit:")
	fmt.Scanln(&initBalance)
	if initBalance <=0{
		return errors.New("Initial deposit can not be zero or negative.")
	}
	fmt.Println("Please set your 4 digit pin:")
	fmt.Scanln(&pin)
	if len(pin) < 4 {
		return errors.New("Your pin is less than 4 digit.")
	} else if len(pin) > 4 {
		return errors.New("Your pin is greater than 4 digit.")
	}
	accNum, err := getAccNum()
	if err != nil {
		return err
	}

	accNumber := strconv.Itoa(accNum)
	newAcc := Accounts{
		HolderName:         name,
		AccountNumber:      accNumber,
		AccountPIN:         pin,
		Balance:            initBalance,
		TransactionHistory: []Transaction{},
	}
	if err := AppendAccount(newAcc); err != nil {
		return err
	}
	fmt.Println("Account created successfully! Your account number is:", accNum)
	return nil
}
func AppendAccount(newAcc Accounts) error {
	accSlice, err := jsonToSlice()
	if err != nil {
		return err
	}
	accSlice = append(accSlice, newAcc)
	if err := saveData(accSlice); err != nil {
		return err
	}
	return nil
}
func getAccNum() (int, error) {
	accSlice, err := jsonToSlice()
	if err != nil {
		return 0, err
	}

	if len(accSlice) == 0 {
		return 12340001, nil
	}

	maxAccNum := 12340000
	for _, acc := range accSlice {
		num, err := strconv.Atoi(acc.AccountNumber)
		if err != nil {
			continue
		}
		if num > maxAccNum {
			maxAccNum = num
		}
	}

	return maxAccNum + 1, nil
}
func Authenticate() (*Accounts, error) {
	var accNum, pin string

	fmt.Println("Please enter your account number:")
	fmt.Scanln(&accNum)

	accSlice, err := jsonToSlice()
	if err != nil {
		return nil, err
	}

	for i := range accSlice {
		if accSlice[i].AccountNumber == accNum {
			fmt.Println("Please enter your PIN:")
			fmt.Scanln(&pin)

			if pin == accSlice[i].AccountPIN {
				return &accSlice[i], nil
			} else {
				return nil, errors.New("wrong PIN")
			}
		}
	}

	return nil, errors.New("account not found")
}
func jsonToSlice() ([]Accounts, error) {
	var accSlice []Accounts

	data, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []Accounts{}, nil
		}
		return []Accounts{}, err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &accSlice); err != nil {
			return []Accounts{}, err
		}
	}
	return accSlice, nil
}
func saveData(accSlice []Accounts) error {
	updatedData, err := json.MarshalIndent(accSlice, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fileName, updatedData, 0644); err != nil {
		return err
	}
	return nil
}
func Deposit(Account *Accounts) error {
	var amount float64
	fmt.Println("Please enter amount you want to deposit:")
	fmt.Scanln(&amount)
	if amount <= 0 {
		return errors.New("invalid amount")
	}
	Account.Balance += amount
	if err := recordTransaction(Account, "Deposit", amount); err != nil {
		return err
	}
	return nil
}
func Withdraw(Account *Accounts) error {
	var amount float64
	fmt.Println("Please enter amount you want to withdraw:")
	fmt.Scanln(&amount)
	if amount <= 0 {
		return errors.New("invalid amount!")
	}
	if Account.Balance < amount {
		return errors.New("Insufficient balance!")
	}
	Account.Balance -= amount
	if err := recordTransaction(Account, "Withdraw", amount); err != nil {
		return err
	}
	return nil
}
func FetchBalance(Account *Accounts) {
	fmt.Println("Your available balance is:", Account.Balance)
}
func Transfer(Account *Accounts) error {
	var toAccount string
	found := false
	accSlice, err := jsonToSlice()
	if err != nil {
		return err
	}
	fmt.Println("Please enter receiver account number:")
	fmt.Scanln(&toAccount)
	if len(toAccount) < 8 {
		return errors.New("Account number is less than 8 digits.")
	} else if toAccount == Account.AccountNumber {
		return errors.New("You can not transfer to your own account.")
	}
	for i := range accSlice {
		if accSlice[i].AccountNumber == toAccount {
			var amount float64
			fmt.Println("Please enter amount you want to transfer:")
			fmt.Scanln(&amount)
			if amount <= 0 {
				return errors.New("Invalid amount")
			} else if amount > Account.Balance {
				return errors.New("Insufficient balance!")
			}
			accSlice[i].Balance += amount
			Account.Balance -= amount
			accSlice[i].TransactionHistory = append(accSlice[i].TransactionHistory, Transaction{
				TransactionTime: time.Now(),
				TransactionType: fmt.Sprintf("Transferred from -> %s", Account.AccountNumber),
				Amount:          amount,
			})
			for j := range accSlice {
				if accSlice[j].AccountNumber == Account.AccountNumber {
					accSlice[j].Balance = Account.Balance
					accSlice[j].TransactionHistory = append(accSlice[j].TransactionHistory, Transaction{
						TransactionTime: time.Now(),
						TransactionType: fmt.Sprintf("Transfer to -> %s", toAccount),
						Amount:          amount,
					})
					Account.TransactionHistory = accSlice[j].TransactionHistory
					break
				}
			}
			if err := saveData(accSlice); err != nil {
				return err
			}
			found = true
			break
		}
	}
	if !found {
		return errors.New("receiver account not found!")
	}
	fmt.Println("Transfer successful!")
	return nil
}
func ChangePin(Account *Accounts) error {
	var pin string
	fmt.Println("Please enter your current pin")
	fmt.Scanln(&pin)
	if Account.AccountPIN == pin {
		var newPin string
		fmt.Println("Please enter your new pin")
		fmt.Scanln(&newPin)
		if len(newPin) < 4 {
			return errors.New("Your new pin is less than 4 digits!")
		} else if len(newPin) > 4 {
			return errors.New("Your new pin is greater than 4 digits!")
		} else {
			accSlice, err := jsonToSlice()
			if err != nil {
				return err
			}
			for i := range accSlice {
				if accSlice[i].AccountNumber == Account.AccountNumber {
					accSlice[i].AccountPIN = newPin
					Account.AccountPIN = newPin
					break
				}
			}
			if err := saveData(accSlice); err != nil {
				return err
			}
			fmt.Println("PIN changed successfully!")
		}
	} else {
		return errors.New("Current PIN is wrong")
	}
	return nil
}
func recordTransaction(Account *Accounts, transactType string, amount float64) error {
	Account.TransactionHistory = append(Account.TransactionHistory, Transaction{
		TransactionTime: time.Now(),
		TransactionType: transactType,
		Amount:          amount,
	})
	accSlice, err := jsonToSlice()
	if err != nil {
		return err
	}
	for i := range accSlice {
		if accSlice[i].AccountNumber == Account.AccountNumber {
			accSlice[i].Balance = Account.Balance
			accSlice[i].TransactionHistory = Account.TransactionHistory
			if err := saveData(accSlice); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
func FetchHistory(Account *Accounts) {
	if len(Account.TransactionHistory) == 0 {
		fmt.Println("No transactions found.")
		return
	}
	for _, t := range Account.TransactionHistory {
		fmt.Printf("Transaction time: %s\nTransaction type: %s\nAmount: %.2f\n\n",
			t.TransactionTime.Format("2006-01-02 15:04:05"), t.TransactionType, t.Amount)
	}
}

func main() {
	menu()
}