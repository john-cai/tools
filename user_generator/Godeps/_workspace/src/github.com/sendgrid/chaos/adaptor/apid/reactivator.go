package apidadaptor

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/now"
	"github.com/sendgrid/chaos/adaptor"
	"github.com/sendgrid/chaos/client"
	"github.com/sendgrid/ln"
)

type Reactivator interface {
	ActivateUserPackage(userID int) *adaptor.AdaptorError
	ActivateUserProfile(userID int) *adaptor.AdaptorError
	AddAdminNote(userID int, userPackage UserPackage) (int, *adaptor.AdaptorError)
	AddBillingNote(adminNoteID int) *adaptor.AdaptorError
	AssignFirstIP(userID int) *adaptor.AdaptorError
	GetPackage(packageID int) (*Package, *adaptor.AdaptorError)
	GetUser(int) (*client.User, *adaptor.AdaptorError)
	CountExternalIP(userID int) (int, *adaptor.AdaptorError)
	GetUserPackage(userID int) (*UserPackage, *adaptor.AdaptorError)
	RemovePunitiveAction(userID int) *adaptor.AdaptorError
	RemoveUserHold(userID int) *adaptor.AdaptorError
	SetUserActive(userID int) *adaptor.AdaptorError
}

const ReactivateActionID = 18

func (a *Adaptor) ActivateUserPackage(userID int) *adaptor.AdaptorError {
	dateStart := time.Now().Format("2006-01-02")
	dateEnd := now.New(time.Now().AddDate(0, 1, 0)).BeginningOfMonth().Format("2006-01-02")
	var result string
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"user_package"},
		"where":     []string{fmt.Sprintf(`{"user_id":"%d"}`, userID)},
		"values":    []string{fmt.Sprintf(`[{"package_status":1}, {"start_date":"%s"}, {"end_date": "%s"}]`, dateStart, dateEnd)},
	}, &result)
	if err != nil {
		ln.Err("Error activating user package", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError(err.Error())
	}
	return nil
}

func (a *Adaptor) SetUserActive(userID int) *adaptor.AdaptorError {
	var apidUserResults int
	err := a.apidClient.DoFunction("setUserActive", url.Values{
		"userid": []string{strconv.Itoa(userID)},
		"active": []string{strconv.Itoa(1)},
	}, &apidUserResults)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return formattedErr
	}

	return nil
}

func (a *Adaptor) ActivateUserProfile(userID int) *adaptor.AdaptorError {
	var apidUserResults int
	err := a.apidClient.DoFunction("editUserProfile", url.Values{
		"userid":            []string{strconv.Itoa(userID)},
		"activated":         []string{strconv.Itoa(1)},
		"is_provision_fail": []string{strconv.Itoa(0)},
		"website_activated": []string{strconv.Itoa(1)},
	}, &apidUserResults)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return formattedErr
	}

	if apidUserResults == 0 {
		return userNotFoundError(userID)
	}

	return nil
}

func (a *Adaptor) RemoveUserHold(userID int) *adaptor.AdaptorError {
	var rowsRemoved int
	err := a.apidClient.DoFunction("removeUserHold", url.Values{
		"userid": []string{strconv.Itoa(userID)},
	}, &rowsRemoved)
	if err != nil {
		formattedErr := adaptor.NewError(err.Error())
		return formattedErr
	}

	return nil
}

func (a *Adaptor) RemovePunitiveAction(userID int) *adaptor.AdaptorError {
	var result string
	currentDate := time.Now().Format("2006-01-02 15:04:05")
	err := a.apidClient.DoFunction("update", url.Values{
		"tableName": []string{"punitive_actions"},
		"where":     []string{fmt.Sprintf(`{"user_id":"%d"}`, userID)},
		"values":    []string{fmt.Sprintf(`[{"active":0}, {"last_action":"%s"}]`, currentDate)},
	}, &result)
	if err != nil {
		ln.Err("Error removing punitive action", ln.Map{"error": err.Error(), "user_id": userID})
		return adaptor.NewError(err.Error())
	}

	return nil
}

func (a *Adaptor) AddAdminNote(userID int, userPackage UserPackage) (int, *adaptor.AdaptorError) {
	currentDate := time.Now().Format("2006-01-02 15:04:05")
	var sqlQuery string
	if userPackage.ID != 0 {
		sqlQuery = fmt.Sprintf(`insert into admin_note (
			user_id,
			agent_id,
			admin_notes_category_id,
			agent_notes,
			package_id,
			package_name,
			created_at
		) values(%d,%d,%d,"Customer updated valid payment information.",%d,"%s","%s")`,
			userID, userID, ReactivateActionID, userPackage.ID, userPackage.PackageName, currentDate)
	} else {
		sqlQuery = fmt.Sprintf(`insert into admin_note (
			user_id,
			agent_id,
			admin_notes_category_id,
			agent_notes,
			created_at
		) values(%d,%d,%d,"Customer updated valid payment information.","%s")`,
			userID, userID, ReactivateActionID, currentDate)
	}

	var adminNoteID int
	err := a.apidClient.DoFunction("executeSql", url.Values{
		"query":    []string{sqlQuery},
		"rw":       []string{"1"},
		"resource": []string{"mail"},
		"insert":   []string{"1"},
	}, &adminNoteID)
	if err != nil {
		ln.Err("Error adding admin note", ln.Map{"error": err.Error(), "user_id": userID})
		return -1, adaptor.NewError(err.Error())
	}

	return adminNoteID, nil
}

func (a *Adaptor) AddBillingNote(adminNoteID int) *adaptor.AdaptorError {
	currentDate := time.Now().Format("2006-01-02 15:04:05")
	var success string
	err := a.apidClient.DoFunction("add", url.Values{
		"tableName": []string{"billing_note"},
		"values": []string{fmt.Sprintf(`[
			{"action_id":"%d"},
			{"admin_note_id":"%d"},
			{"created_at":"%s"},
			{"updated_at":"%s"},
		]`, ReactivateActionID, adminNoteID, currentDate, currentDate)},
	}, &success)
	if err != nil {
		ln.Err("Error adding billing note", ln.Map{"error": err.Error(), "admin_note_id": adminNoteID})
		return adaptor.NewError(err.Error())
	}

	return nil
}
