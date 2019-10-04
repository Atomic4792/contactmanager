package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type formPostData struct {
	ID          string `form:"contactID"sql:"id" json:"id"` //I don't think conactID is getting any values
	FirstName   string `form:"firstName"sql:"first_name" json:"first_name"`
	LastName    string `form:"lastName"sql:"last_name" json:"last_name"`
	Phone       int    `form:"phone"sql:"phone" json:"phone"`
	OfficePhone int    `form:"officePhone"sql:"office_phone" json:"office_phone"`
	City        string `form:"city"sql:"city" json:"city"`
	State       string `form:"state"sql:"state" json:"state"`
	Zip         string `form:"zip"sql:"zip" json:"zip"`
}

func NewFormPostData() formPostData {
	return formPostData{}
}

type ContactInfo struct {
	ID          string `sql:"id" json:"id"`
	FirstName   string `sql:"first_name" json:"first_name"`
	LastName    string `sql:"last_name" json:"last_name"`
	Phone       int    `sql:"phone" json:"phone"`
	OfficePhone int    `sql:"office_phone" json:"office_phone"`
	City        string `sql:"city" json:"city"`
	State       string `sql:"state" json:"state"`
	Zip         string `sql:"zip" json:"zip"`
	Enabled     bool   `sql:"enabled" json:"enabled"`
}

func NewContact() ContactInfo {
	contact := ContactInfo{}
	return contact
}

func (ac *appContext) ShowIndex(c *gin.Context) {
	ac.Log.Msg(1, "in context show index")

	query :=
		` select 
			id, first_name , last_name, phone, office_phone, city, state, zip, enabled 
		from contacts 
		where 
			enabled
			order by 3,2`

	rows, err := ac.DB.Query(query)
	if check := ac.DBErrorCheck(err, query, c); check == false {
		ac.Log.Msg(1, "db error")
		return
	}
	defer rows.Close()

	var contacts []ContactInfo

	ac.Log.Msg(1, "before assignment")
	for rows.Next() {
		contact := NewContact()

		err := rows.Scan(&contact.ID, &contact.FirstName, &contact.LastName, &contact.Phone, &contact.OfficePhone,
			&contact.City, &contact.State, &contact.Zip, &contact.Enabled)
		if err != nil {
			ac.Log.Msg(3, fmt.Sprintf("Error scanning row: %s", err.Error()))
		}

		ac.Log.Msg(1, fmt.Sprintf("%#v", contact))

		contacts = append(contacts, contact)
	}

	c.HTML(http.StatusOK, "main/index", gin.H{
		"contacts": contacts,
	})

}

/*func (ac *appContext) uploadContact(c *gin.Context) {
	var form formPostData

	if err := c.Bind(&form); err != nil {
		fmt.Println("bind error" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": "ok"})
	fmt.Printf("%+v", form)
	query := "insert into contacts (first_name, last_name, phone, office_phone, " +
		"city, state, zip) " +
		"values ($1, $2, $3, $4, $5, $6, $7)"
	_, err := ac.DB.Exec(query, &form.FirstName, &form.LastName, &form.Phone,
		&form.OfficePhone, &form.City, &form.State, &form.Zip)
	ac.DBErrorCheck(err, query, c)

} */
func (ac *appContext) editContact(c *gin.Context) {
	form := NewFormPostData()

	if err := c.Bind(&form); err != nil {
		ac.Log.Msg(3, fmt.Sprintf("bind error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `
		select 
			id, first_name, last_name, phone, office_phone, city, 
state, zip 
		from contacts 
		where
			id = $1
	`
	row := ac.DB.QueryRow(query, &form.ID)
	err := row.Scan(&form.ID, &form.FirstName, &form.LastName, &form.Phone, &form.OfficePhone, &form.City, &form.State, &form.Zip)
	if check := ac.DBErrorCheck(err, query, c); check == false {
		ac.AbortMsg(http.StatusInternalServerError, err, c)
	}
	c.JSON(http.StatusOK, gin.H{
		"ID":          form.ID,
		"FirstName":   form.FirstName,
		"LastName":    form.LastName,
		"Phone":       form.Phone,
		"OfficePhone": form.OfficePhone,
		"City":        form.City,
		"State":       form.State,
		"Zip":         form.Zip,
	})
}
func (ac *appContext) saveContact(c *gin.Context) {
	form := NewFormPostData()

	if err := c.Bind(&form); err != nil {
		ac.Log.Msg(3, fmt.Sprintf("bind error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if form.ID == "" {
		query := `insert into contacts (first_name, last_name, phone, office_phone, city, state, zip)
values ($1, $2, $3, $4, $5, $6, $7) returning id; `
		row := ac.DB.QueryRow(query, &form.FirstName, &form.LastName, &form.Phone, &form.OfficePhone, &form.City, &form.State, &form.Zip)
		err := row.Scan(&form.ID)
		if check := ac.DBErrorCheck(err, query, c); check == false {
			ac.AbortMsg(http.StatusInternalServerError, err, c)
		}
	} else {
		query := `
			update contacts set 
				first_name = $1,
				last_name = $2,
				phone = $3,
				office_phone = $4,
				city= $5,
				state= $6,
				zip= $7

			where
				id = $8
		`
		res, err := ac.DB.Exec(query, &form.FirstName, &form.LastName, &form.Phone, &form.OfficePhone, &form.City, &form.State,  &form.Zip, &form.ID)
		if check := ac.DBErrorCheck(err, query, c); check == false {
			ac.AbortMsg(http.StatusInternalServerError, err, c)
			return
		}
		ra, _ := res.RowsAffected()

		ac.Log.Msg(0, fmt.Sprintf("rows affected [ %d ]", ra))




	}

}
func (ac *appContext) deleteContact(c *gin.Context) {
	form := NewFormPostData()

	if err := c.Bind(&form); err != nil {
		ac.Log.Msg(3, fmt.Sprintf("bind error: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := `
		delete from contacts
		where
			id = $1
`

	res, err := ac.DB.Exec(query, &form.ID)
	if check := ac.DBErrorCheck(err, query, c); check == false {
		ac.AbortMsg(http.StatusInternalServerError, err, c)
		return
	}
	ra, err := res.RowsAffected()
	ac.Log.Msg(0, fmt.Sprintf("rows affected [ %d ]", ra))

}
