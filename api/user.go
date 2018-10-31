package api

import (
	"context"
	"net/http"

	"github.com/cyrinux/monitor2call/database"
	"github.com/cyrinux/monitor2call/models"        //models
	"github.com/cyrinux/monitor2call/notifications" //notifications
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// PostUser godoc
// @Summary Create a user
// @Description create a user
// @Accept application/json
// @Produce json
// @Param name path string true "id"
// @Param language query string true "language"
// @Param email path string true "email"
// @Param pushover_api_key path string true "pushover api key"
// @Param phone path string true "phone"
// @Param tags path array true "tag list"
// @Failure 422 {string} string
// @Success 201 {object} models.User
// @Success 409 {object} models.User
// @Router /api/v1/users [post]
func PostUser(c *gin.Context) {
	db := database.InitDb()

	var user models.User

	// // always return 400 if used BUG!: https://github.com/gin-gonic/gin/pull/1047
	// if err := c.ShouldBindJSON(&user); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	c.BindJSON(&user)

	if err := user.Validate(); err != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
		return
	}

	// prepare the message
	user = user.Prepare()

	// check is pushover key is valid
	if err := notifications.CheckPushoverUserKey(user.PushoverAPIKey); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// test if user allready exists
	if row, err := db.Get(context.TODO(), user.Name); err == nil {
		if err = row.ScanDoc(&user); err == nil {
			log.Print(err)
			c.JSON(http.StatusConflict, gin.H{"success": user})
			return
		}
	}

	// set docID with username
	user.ID = user.Name

	if _, err := db.Put(context.TODO(), user.Name, &user); err != nil {
		log.Print(err)

	}

	if row, err := db.Get(context.TODO(), user.Name); err == nil {
		if err = row.ScanDoc(&user); err == nil {
			c.JSON(http.StatusCreated, gin.H{"success": user})
			return
		}
	}
}

// GetUsers fetch all users list
// @Summary Get all users
// @Description Get all monitor alerts
// @Accept json
// @Produce json
// @Success 200 {object} models.User[]
// @Failure 404 {string} string
// @Router /api/v1/users [get]
func GetUsers(c *gin.Context) {
	db := database.InitDb()

	var user models.User
	var users []models.User

	rows, err := db.Query(context.TODO(), "_design/users", "_view/all")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No users found"})
		return
	}

	for rows.Next() {
		id := rows.ID()
		row, err := db.Get(context.TODO(), id)
		if err != nil {
			log.Printf("Can't get user %s", id)
		}
		if err := row.ScanDoc(&user); err != nil {
			log.Print(err)
		} else {
			users = append(users, user)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": &users})
	return
}

// GetUser godoc
// @Summary Get alert by id
// @Description get struct array by ID
// @Param id path int true "4"
// @Success 200 {object} models.User
// @Success 404 {string} string
// @Router /api/v1/users/{id} [get]
func GetUser(c *gin.Context) {
	db := database.InitDb()

	var user models.User
	c.Bind(&user)

	name := c.Params.ByName("id")

	row, err := db.Get(context.TODO(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		c.Abort()
		return
	}

	if err = row.ScanDoc(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Can't read user"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": &user})
	return
}

// // Edit a user
// func EditUser(c *gin.Context) {
// 	db := InitDb()

// 	// Récupération de l'id dans une variable
// 	id := c.Params.ByName("id")
// 	var user User
// 	// SELECT * FROM users WHERE id = id;
// 	db.First(&user, id)

// 	if user.Name != "" {
// 		if user.ID != 0 {
// 			var json User
// 			c.Bind(&json)

// 			// UPDATE users SET name='json.Name' WHERE id = user.ID;
// 			db.Save(&user)
// 			// Affichage des données modifiées
// 			c.JSON(200, gin.H{"success": user})
// 		} else {
// 			// Affichage de l'erreur
// 			c.JSON(404, gin.H{"error": "User not found"})
// 		}

// 	} else {
// 		// Affichage de l'erreur
// 		c.JSON(422, gin.H{"error": "Fields are empty"})
// 	}
// }

// DeleteUser delete an user
// @Summary Delete an user
// @Description Delete an user with id and rev
// @Param id path string true "toto"
// @Param _rev query string true "1-f91a0d21bd7476a9cd8853da0846223d"
// @Accept application/json
// @Produce json
// @Failure 404 {string} string
// @Success 200 {string} string
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	db := database.InitDb()

	var user models.User
	c.BindJSON(&user)

	name := c.Params.ByName("id")
	rev := user.Rev
	newRev, err := db.Delete(context.TODO(), name, rev)
	if err != nil {
		log.Errorf("User %s not found", name)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "User deleted " + name + ", new revision is " + newRev})
}

// OptionsUser set allowed method
func OptionsUser(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Next()
}
