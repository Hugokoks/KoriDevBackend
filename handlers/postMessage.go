package handlers

import (
	"koridev/mail"
	"koridev/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PostMessage(c *gin.Context) {
	var form models.ContactForm

	if err := c.ShouldBindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"message": "Invalid input",
		})
		return
	}

	////SEND Email
	if err := mail.SendMailResend(form.Name, form.Email, form.Message); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"valid":   false,
			"message": "The message failed to send. Please try again later.",
		})

		return

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "The message sent successfully.",
		"valid":   true,
	})
}
