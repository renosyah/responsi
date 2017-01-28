package main

import "html/template"
import "net/http"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "github.com/skratchdot/open-golang/open"
import "os"
import "fmt"
import "io"

var username, pass, db string

func connect() *sql.DB {
	var db, err = sql.Open("mysql", username+":"+pass+"@/"+db)
	err = db.Ping()
	if err != nil {
		fmt.Println("database tidak bisa dihubungi")
		os.Exit(0)

	}
	return db

}

func input_db_akses() {
	var bs string
	fmt.Print("masukkan username mysql : ")
	fmt.Scanln(&username)
	fmt.Print("masukkan password mysql : ")
	fmt.Scanln(&pass)
	fmt.Print("masukkan database mysql : ")
	fmt.Scanln(&db)
	fmt.Print("Browser : ")
	fmt.Scanln(&bs)

	db := connect()
	defer db.Close()
	open.RunWith("http://localhost:8080/", bs)

}

type maba struct { // isi tidak boleh di deklarasikan sembarangan
	Number  int
	Id      string
	Name    string
	Text    string
	Profile string
}

func tampil(res http.ResponseWriter, req *http.Request) {
	db := connect()
	defer db.Close()

	rows, _ := db.Query("select * from mhs order by nim asc")

	var nim, nama, status, foto string
	no := 1

	type mahasiswa []maba  // interface ke map
	var data_mhs mahasiswa // map baru

	for rows.Next() {

		rows.Scan(&nim, &nama, &status, &foto)
		data := maba{
			Number:  no,
			Id:      nim,
			Name:    nama,
			Text:    status,
			Profile: foto,
		}
		no++
		data_mhs = append(data_mhs, data) // map diisi
	}
	t, _ := template.ParseFiles("tabel.html")

	t.Execute(res, data_mhs)
}

func isi_data(res http.ResponseWriter, req *http.Request) {
	db := connect()
	defer db.Close()

	nim := req.FormValue("nim")
	nama := req.FormValue("nama")
	status := req.FormValue("status")
	foto, header, foto_error := req.FormFile("foto")
	var file_foto string

	var path_foto = "data/profil/" + nim + "/"
	os.MkdirAll(path_foto, 0777)

	switch foto_error {
	case nil:
		file_foto = header.Filename
		out, _ := os.Create(path_foto + file_foto)
		io.Copy(out, foto)
		defer out.Close()

	case http.ErrMissingFile:

		file_foto = "../../default.jpg"
	default:
		file_foto = "../../default.jpg"
	}

	db.Exec("insert into mhs value (?,?,?,?)", nim, nama, status, file_foto)

	http.Redirect(res, req, "/", 301)
}
func hapus(res http.ResponseWriter, req *http.Request) {

	db := connect()
	defer db.Close()

	nim := req.FormValue("hapus")
	var foto string
	db.QueryRow("select foto from mhs where nim=?", nim).Scan(&foto)
	if foto != "../../default.jpg" {

		os.RemoveAll("data/profil/" + nim + "/" + foto)
		os.Remove("data/profil/" + nim + "/")
	}
	if foto == "../../default.jpg" {
		os.RemoveAll("data/profil/" + nim + "/")
	}
	db.Exec("delete from mhs where nim=?", nim)

	http.Redirect(res, req, "/", 301)
}
func ubah(res http.ResponseWriter, req *http.Request) {

	db := connect()
	defer db.Close()

	nim := req.FormValue("nim")
	nama := req.FormValue("nama")
	status := req.FormValue("status")
	id := req.FormValue("id")

	if nim != "" {
		db.Exec("update mhs set nim=? where nim=?", nim, id)
	}
	if nama != "" {
		db.Exec("update mhs set nama=? where nim=?", nama, id)
	}
	if status != "" {
		db.Exec("update mhs set status=? where nim=?", status, id)
	}
	http.Redirect(res, req, "/", 301)

}

func ubah_foto(res http.ResponseWriter, req *http.Request) {

	db := connect()
	defer db.Close()

	id := req.FormValue("id")
	foto, foto_baru, _ := req.FormFile("foto")

	var file_foto_baru, foto_lama string
	file_foto_baru = foto_baru.Filename

	db.QueryRow("select foto from mhs where nim=?", id).Scan(&foto_lama)

	if foto_lama != "../../default.jpg" {
		os.Remove("data/profil/" + id + "/" + foto_lama)
		out, _ := os.Create("data/profil/" + id + "/" + file_foto_baru)

		io.Copy(out, foto)
		defer out.Close()
	}
	if foto_lama == "../../default.jpg" {
		out, _ := os.Create("data/profil/" + id + "/" + file_foto_baru)

		io.Copy(out, foto)
		defer out.Close()
	}

	db.Exec("update mhs set foto=? where nim=?", file_foto_baru, id)

	http.Redirect(res, req, "/", 301)
}
func main() {
	input_db_akses()
	http.HandleFunc("/", tampil)
	http.HandleFunc("/isi_data", isi_data)
	http.HandleFunc("/hapus", hapus)
	http.HandleFunc("/ubah", ubah)
	http.HandleFunc("/ubah_foto", ubah_foto)
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))
	fmt.Println("running server via localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
