package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/felipemendozag/api-sis-operativos/db"
)

type AttendanceReports struct {
	TypeStatus string `json:"type_status"`
	Count      int    `json:"count"`
}
type AttendanceForUser struct {
	Status   string `json:"status"`
	FullName string `json:"full_name"`
	Count    string `json:"count"`
}
type AttendanceForType struct {
	Status string `json:"status"`
	Count  string `json:"count"`
}
type AttendanceForDate struct {
	DateFormat string `json:"date_format"`
	Count      int    `json:"count"`
}

func GetReports(w http.ResponseWriter, r *http.Request) {
	var filter DateFilter

	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dateStart := filter.DateStart + " 00:00:00"
	dateEnd := filter.DateEnd + " 23:59:59"
	query_sql := `
		select 
		CASE when tb.is_read then 'Leidos' else 'No Leidos' end as type_status, 
		count(tb.*) from (
		select
		u2.first_name, u2.last_name, n.id, n."type" , n.is_read, n.created_at, n.read_at 
		from notifications n 
		inner join users u2 on u2.id = n.user_id 
		where n.created_at between ? and ?
		) as tb
		group by tb.is_read;
	`
	query_sql_2 := `
		select tb.status, tb.full_name, count(tb.status) from (
			select
			a.user_id , a.status, a."date",
			concat(u.first_name, ' ', u.last_name) as full_name 
			from attendances a 
			inner join users u on u.id = a.user_id 
			where
			a.created_at between ? and ?
		) as tb
		group by tb.full_name, tb.status;
	`
	query_sql_3 := `
	select a.status, count(*) FROM attendances a
		WHERE a.created_at between ? and ?
		group by a.status ;
	`
	query_sql_4 := `
	select tb.date_format, count(*) as "count" from (
		select
		a.id , a.user_id , DATE(a."date") as date_format, a.status
		from attendances a
		WHERE a.created_at between ? and ?
	) as tb
	group by tb.date_format order by tb.date_format desc;
	`

	var resultAttendanceForUser []AttendanceForUser
	var resultAttendance []AttendanceReports
	var resultAttendanceType []AttendanceForType
	var resultAttendanceDate []AttendanceForDate
	if err := db.DB.Raw(query_sql, dateStart, dateEnd).Scan(&resultAttendance).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err2 := db.DB.Raw(query_sql_2, dateStart, dateEnd).Scan(&resultAttendanceForUser).Error; err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	if err3 := db.DB.Raw(query_sql_3, dateStart, dateEnd).Scan(&resultAttendanceType).Error; err3 != nil {
		http.Error(w, err3.Error(), http.StatusInternalServerError)
		return
	}
	if err4 := db.DB.Raw(query_sql_4, dateStart, dateEnd).Scan(&resultAttendanceDate).Error; err4 != nil {
		http.Error(w, err4.Error(), http.StatusInternalServerError)
		return
	}
	if resultAttendance == nil {
		resultAttendance = []AttendanceReports{}
	}
	if resultAttendanceForUser == nil {
		resultAttendanceForUser = []AttendanceForUser{}
	}
	if resultAttendanceType == nil {
		resultAttendanceType = []AttendanceForType{}
	}
	if resultAttendanceDate == nil {
		resultAttendanceDate = []AttendanceForDate{}
	}
	response := map[string]interface{}{
		"attendance_reports":       resultAttendance,
		"attendance_reports_users": resultAttendanceForUser,
		"attendance_reports_type":  resultAttendanceType,
		"attendance_reports_date":  resultAttendanceDate,
	}
	fmt.Println(query_sql)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
