package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"infotecs/internal/models"
	"infotecs/internal/service"
	"io"
	"net/http"
	"time"
)

func PostNewWallet(c *gin.Context) {
	if err := service.Service.Repository.BeginTransaction(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer service.Service.Repository.RollbackTransaction()

	wallet := &models.Wallet{
		Balance: 100,
	}
	var err error
	wallet.Id, err = service.Service.Repository.SaveWallet(wallet)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	service.Service.Repository.CommitTransaction()
	c.IndentedJSON(http.StatusCreated, wallet)
}

func PostSendMoney(c *gin.Context) {
	id := c.Param("id")

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "invalid request json body")
		return
	}
	defer c.Request.Body.Close()

	var transaction models.Transaction
	if err := json.Unmarshal(data, &transaction); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	if transaction.Amount < 0 {
		c.IndentedJSON(http.StatusBadRequest, "amount of money can't be negative")
		return
	}
	transaction.Time = &models.CustomTime{Time: time.Now()}
	transaction.From = id

	wallet1, err := service.Service.Repository.GetWallet(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, "wallet not found")
		} else {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	wallet2, err := service.Service.Repository.GetWallet(transaction.To)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusBadRequest, "wallet not found")
		} else {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}
		return
	}

	if wallet1.Balance-transaction.Amount < 0 {
		c.IndentedJSON(http.StatusBadRequest, "the wallet does not have enough money for the transaction")
		return
	}
	wallet1.Balance -= transaction.Amount
	wallet2.Balance += transaction.Amount

	if err := service.Service.Repository.BeginTransaction(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer service.Service.Repository.RollbackTransaction()

	if err := service.Service.Repository.UpdateWallet(wallet1); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	if err := service.Service.Repository.UpdateWallet(wallet2); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := service.Service.Repository.SaveTransaction(&transaction); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	service.Service.Repository.CommitTransaction()
	c.IndentedJSON(http.StatusOK, transaction)

}

func GetTransactionHistory(c *gin.Context) {
	id := c.Param("id")

	_, err := service.Service.Repository.GetWallet(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, "wallet not found")
		} else {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}
		return
	}

	transactions, err := service.Service.Repository.GetTransactionHistory(id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}

	c.IndentedJSON(http.StatusOK, transactions)

}

func GetWallet(c *gin.Context) {
	id := c.Param("id")

	wallet, err := service.Service.Repository.GetWallet(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, "wallet not found")
		} else {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.IndentedJSON(http.StatusOK, wallet)

}
