package edumasbackend

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/whatsauth/watoken"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ReturnStruct(DataStuct any) string {
	jsondata, _ := json.Marshal(DataStuct)
	return string(jsondata)
}

// <--- ini Login & Register User --->

func Register(Mongoenv, dbname string, r *http.Request) string {
	resp := new(Credential)
	userdata := new(User)
	resp.Status = false
	conn := SetConnection(Mongoenv, dbname)
	err := json.NewDecoder(r.Body).Decode(&userdata)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		resp.Status = true
		hash, err := HashPass(userdata.Password)
		if err != nil {
			resp.Message = "Gagal Hash Password" + err.Error()
		}
		InsertUserdata(conn, userdata.Username, userdata.Role, hash)
		resp.Message = "Berhasil Input data"
	}
	response := ReturnStringStruct(resp)
	return response
}

func RegisterNew(Mongoenv, dbname string, r *http.Request) string {
	resp := new(Credential)
	userdata2 := new(UserNew)
	resp.Status = false
	conn := SetConnection(Mongoenv, dbname)
	err := json.NewDecoder(r.Body).Decode(&userdata2)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		resp.Status = true
		hash, err := HashPass(userdata2.Password)
		if err != nil {
			resp.Message = "Gagal Hash Password" + err.Error()
		}
		InsertUserdataNew(conn, userdata2.Username , userdata2.Notelp, userdata2.Role, hash)
		resp.Message = "Berhasil Input data"
	}
	response := ReturnStringStruct(resp)
	return response
}

func LoginUserNew(Privatekey, MongoEnv, dbname, Colname string, r *http.Request) string {
	var resp Credential
	mconn := SetConnection(MongoEnv, dbname)
	var datauser2 UserNew
	err := json.NewDecoder(r.Body).Decode(&datauser2)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		if IsPasswordValidUserNew(mconn, Colname, datauser2) {
			tokenstring, err := watoken.Encode(datauser2.Username, os.Getenv(Privatekey))
			if err != nil {
				resp.Message = "Gagal Encode Token : " + err.Error()
			} else {
				resp.Status = true
				resp.Message = "Selamat Datang User"
				resp.Token = tokenstring
			}
		} else {
			resp.Message = "Password Salah"
		}
	}
	return GCFReturnStruct(resp)
}


func Login(Privatekey, MongoEnv, dbname, Colname string, r *http.Request) string {
	var resp Credential
	mconn := SetConnection(MongoEnv, dbname)
	var datauser User
	err := json.NewDecoder(r.Body).Decode(&datauser)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		if IsPasswordValid(mconn, Colname, datauser) {
			tokenstring, err := watoken.Encode(datauser.Username, os.Getenv(Privatekey))
			if err != nil {
				resp.Message = "Gagal Encode Token : " + err.Error()
			} else {
				resp.Status = true
				resp.Message = "Selamat Datang User"
				resp.Token = tokenstring
			}
		} else {
			resp.Message = "Password Salah"
		}
	}
	return GCFReturnStruct(resp)
}

//GetAllUser
func GetAllDataUser(PublicKey, MongoEnv, dbname, colname string, r *http.Request) string {
	req := new(Response)
	conn := SetConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")
	if tokenlogin == "" {
		req.Status = false
		req.Message = "Header Login Not Found"
	} else {
		// Dekode token untuk mendapatkan
		_, err := DecodeGetUser(os.Getenv(PublicKey), tokenlogin)
		if err != nil {
			req.Status = false
			req.Message = "Data Tersebut tidak ada" + tokenlogin
		} else {
			// Langsung ambil data user
			datauser := GetAllUser(conn, colname)
			if datauser == nil {
				req.Status = false
				req.Message = "Data User tidak ada"
			} else {
				req.Status = true
				req.Message = "Data User berhasil diambil"
				req.Data = datauser
			}
		}
	}
	return ReturnStringStruct(req)
}

//Get satu data User
func GetOneDataUser(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	resp := new(Credential)
	userdata := new(UserNew)
	resp.Status = false
	err := json.NewDecoder(r.Body).Decode(&userdata)

	id := r.URL.Query().Get("_id")
	if id == "" {
		resp.Message = "Missing '_id' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		resp.Message = "Invalid '_id' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	userdata.ID = ID

	// Menggunakan fungsi GetProdukFromID untuk mendapatkan data produk berdasarkan ID
	userdata, err = GetUserFromID(mconn, collectionname, ID)
	if err != nil {
		resp.Message = err.Error()
		return GCFReturnStruct(resp)
	}

	resp.Status = true
	resp.Message = "Get Data Berhasil"
	resp.Datas = []UserNew{*userdata}

	return GCFReturnStruct(resp)
}

func GCFFindUserByName(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var datauser User
	err := json.NewDecoder(r.Body).Decode(&datauser)
	if err != nil {
		return err.Error()
	}

	// Jika username kosong, maka respon "false" dan data tidak ada
	if datauser.Username == "" {
		return "false"
	}

	// Jika ada username, mencari data pengguna
	user := FindUserUser(mconn, collectionname, datauser)

	// Jika data pengguna ditemukan, mengembalikan data pengguna dalam format yang sesuai
	if user != (User{}) {
		return GCFReturnStruct(user)
	}

	// Jika tidak ada data pengguna yang ditemukan, mengembalikan "false" dan data tidak ada
	return "false"
}

//Function Admin
func LoginAdmin(Privatekey, MongoEnv, dbname, Colname string, r *http.Request) string {
	var resp Credential
	mconn := SetConnection(MongoEnv, dbname)
	var dataadmin Admin
	err := json.NewDecoder(r.Body).Decode(&dataadmin)
	if err != nil {
		resp.Message = "error parsing application/json: " + err.Error()
	} else {
		if IsPasswordValidAdmin(mconn, Colname, dataadmin) {
			tokenstring, err := watoken.Encode(dataadmin.Username, os.Getenv(Privatekey))
			if err != nil {
				resp.Message = "Gagal Encode Token : " + err.Error()
			} else {
				resp.Status = true
				resp.Message = "Selamat Datang Admin"
				resp.Token = tokenstring
			}
		} else {
			resp.Message = "Password Salah"
		}
	}
	return GCFReturnStruct(resp)
}

// return struct
func GCFReturnStruct(DataStuct any) string {
	jsondata, _ := json.Marshal(DataStuct)
	return string(jsondata)
}

func ReturnStringStruct(Data any) string {
	jsonee, _ := json.Marshal(Data)
	return string(jsonee)
}

// Insert Report post 
func GCFInsertReport(publickey, MONGOCONNSTRINGENV, dbname, colluser, collreport string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var userdata User
	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		response.Message = "Missing Login in headers"
	} else {
		// Process the request with the "Login" token
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		userdata.Username = checktoken
		if checktoken == "" {
			response.Message = "Invalid token"
		} else {
			user2 := FindUser(mconn, colluser, userdata)
			if user2.Role == "user" {
				var datareport Report
				err := json.NewDecoder(r.Body).Decode(&datareport)
				if err != nil {
					response.Message = "Error parsing application/json: " + err.Error()
				} else {
					insertReport(mconn, collreport, Report{
						Nik:     		datareport.Nik,
						Nama:			datareport.Nama,
						Title:       	datareport.Title,
						Description: 	datareport.Description,
						DateOccurred: 	datareport.DateOccurred,
						Image:       	datareport.Image,
						Status:      	datareport.Status,
						PihakTerkait: 	datareport.PihakTerkait,
						Tanggapan: 		datareport.Tanggapan,
					})
					response.Status = true
					response.Message = "Berhasil Insert Report"
				}
			} else {
				response.Message = "Anda tidak bisa Insert data karena bukan User"
			}
		}
	}
	return GCFReturnStruct(response)
}


//Delete Report For Admin
func GCFDeleteReportForAdmin(publickey, MONGOCONNSTRINGENV, dbname, colladmin, collreport string, r *http.Request) string {

	var respon Credential
	respon.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var admindata Admin

	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		respon.Message = "Missing Login in headers"
	} else {
		// Process the request with the "Login" token
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		admindata.Username = checktoken
		if checktoken == "" {
			respon.Message = "Invalid token"
		} else {
			admin2 := FindAdmin(mconn, colladmin, admindata)
			if admin2.Role == "admin" {
				var datareport Report
				err := json.NewDecoder(r.Body).Decode(&datareport)
				if err != nil {
					respon.Message = "Error parsing application/json: " + err.Error()
				} else {
					DeleteReport(mconn, collreport, datareport)
					respon.Status = true
					respon.Message = "Berhasil Delete Report"
				}
			} else {
				respon.Message = "Anda tidak bisa Delete data karena bukan Admin"
			}
		}
	}
	return GCFReturnStruct(respon)
}


// Update report for admin
func GCFUpdateReportForAdmin(publickey, MONGOCONNSTRINGENV, dbname, colladmin, collreport string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var admindata Admin

	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		response.Message = "Missing Login in Headers"
	} else {
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		admindata.Username = checktoken
		if checktoken == "" {
			response.Message = "Invalid token"
		} else {
			admin2 := FindAdmin(mconn, colladmin, admindata)
			if admin2.Role == "admin" {
				var datareport Report
				err := json.NewDecoder(r.Body).Decode(&datareport)
				if err != nil {
					response.Message = "Error parsing application/json: " + err.Error()

				} else {
					UpdatedReport(mconn, collreport, bson.M{"id": datareport.ID}, datareport)
					response.Status = true
					response.Message = "Berhasil Update Report"
					GCFReturnStruct(CreateResponse(true, "Success Update Report", datareport))
				}
			} else {
				response.Message = "Anda tidak bisa Update data karena bukan Admin"
			}

		}
	}
	return GCFReturnStruct(response)
}

// Update data user for admin
func GCFUpdateUserForAdmin(publickey, MONGOCONNSTRINGENV, dbname, colladmin, colluser string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var admindata Admin

	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		response.Message = "Missing Login in Headers"
	} else {
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		admindata.Username = checktoken
		if checktoken == "" {
			response.Message = "Invalid token"
		} else {
			admin2 := FindAdmin(mconn, colladmin, admindata)
			if admin2.Role == "admin" {
				var datauser UserNew
				err := json.NewDecoder(r.Body).Decode(&datauser)
				if err != nil {
					response.Message = "Error parsing application/json: " + err.Error()

				} else {
					UpdatedUser(mconn, colluser, bson.M{"username": datauser.Username}, datauser)
					response.Status = true
					response.Message = "Berhasil Update User"
					GCFReturnStruct(CreateResponse(true, "Success Update User", datauser))
				}
			} else {
				response.Message = "Anda tidak bisa Update data karena bukan Admin"
			}

		}
	}
	return GCFReturnStruct(response)
}

// Update data user for User
func GCFUpdateUserForUser(publickey, MONGOCONNSTRINGENV, dbname, colluser string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var userdata UserNew

	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		response.Message = "Missing Login in Headers"
	} else {
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		userdata.Username = checktoken
		if checktoken == "" {
			response.Message = "Invalid token"
		} else {
			user2 := FindUserNew(mconn, colluser, userdata)
			if user2.Role == "user" {
				var datauser UserNew	
				err := json.NewDecoder(r.Body).Decode(&datauser)
				if err != nil {
					response.Message = "Error parsing application/json: " + err.Error()

				} else {
					UpdatedUser(mconn, colluser, bson.M{"username": datauser.Username}, datauser)
					response.Status = true
					response.Message = "Berhasil Update User"
					GCFReturnStruct(CreateResponse(true, "Success Update User", datauser))
				}
			} else {
				response.Message = "Anda tidak bisa Update data karena bukan User"
			}

		}
	}
	return GCFReturnStruct(response)
}

func UpdateUserByUser(publickey, mongoenv, dbname, collname string, r *http.Request) string {
	var response Credential
	response.Status = false

	// Establish MongoDB connection
	mconn := SetConnection(mongoenv, dbname)

	// Decode user data from the request body
	var auth UserNew
	var datauser UserNew
	err := json.NewDecoder(r.Body).Decode(&datauser)

	// Check for JSON decoding errors
	if err != nil {
		response.Message = "Error parsing application/json: " + err.Error()
		return ReturnStruct(response)
	}

	// Get token and perform basic token validation
	header := r.Header.Get("Login")
	if header == "" {
		response.Message = "Header login tidak ditemukan"
		return ReturnStruct(response)
	}

	// Decode token to get username and role
	tokenusername := DecodeGetUsername(os.Getenv(publickey), header)
	tokenrole := DecodeGetRole(os.Getenv(publickey), header)
	auth.Username = tokenusername

	// Check if decoding was successful
	if tokenusername == "" || tokenrole == "" {
		response.Message = "Hasil decode tidak ditemukan"
		return ReturnStruct(response)
	}

	// Check if the user account exists
	if !usernameExists(mongoenv, dbname, auth) {
		response.Message = "Akun tidak ditemukan"
		return ReturnStruct(response)
	}

	// Check if the user has user privileges
	if tokenrole != "user" {
		response.Message = "Anda tidak memiliki akses"
		return ReturnStruct(response)
	}

	// Check if the username parameter is provided
	if datauser.Username == "" {
		response.Message = "Parameter dari function ini adalah username"
		return ReturnStruct(response)
	}

	// Check if the user to be edited exists
	if !usernameExists(mongoenv, dbname, datauser) {
		response.Message = "Akun yang ingin diedit tidak ditemukan"
		return ReturnStruct(response)
	}

	// Hash the user's password if provided
	if datauser.Password != "" {
		hash, hashErr := HashPass(datauser.Password)
		if hashErr != nil {
			response.Message = "Gagal Hash Password: " + hashErr.Error()
			return ReturnStruct(response)
		}
		datauser.Password = hash
	} else {
		// Retrieve user details
		user := FindUserNews(mconn, collname, datauser)
		datauser.Password = user.Password
	}

	// Perform user update
	EditUser(mconn, collname, datauser)

	response.Status = true
	response.Message = "Berhasil update " + datauser.Username + " dari database"
	return ReturnStruct(response)
}

func GCFDeleteUserForAdmin(publickey, MONGOCONNSTRINGENV, dbname, colladmin, colluser string, r *http.Request) string {

	var respon Credential
	respon.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var admindata Admin

	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		respon.Message = "Missing Login in headers"
	} else {
		// Process the request with the "Login" token
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		admindata.Username = checktoken
		if checktoken == "" {
			respon.Message = "Invalid token"
		} else {
			admin2 := FindAdmin(mconn, colladmin, admindata)
			if admin2.Role == "admin" {
				var datauser UserNew
				err := json.NewDecoder(r.Body).Decode(&datauser)
				if err != nil {
					respon.Message = "Error parsing application/json: " + err.Error()
				} else {
					Deleteuser(mconn, colluser, datauser)
					respon.Status = true
					respon.Message = "Berhasil Delete User"
				}
			} else {
				respon.Message = "Anda tidak bisa Delete data karena bukan Admin"
			}
		}
	}
	return GCFReturnStruct(respon)
}

// get all report
func GCFGetAllReport(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	datareport := GetAllReport(mconn, collectionname)
	if datareport != nil {
		return GCFReturnStruct(CreateResponse(true, "success Get All Report", datareport))
	} else {
		return GCFReturnStruct(CreateResponse(false, "Failed Get All Report", datareport))
	}
}

//Get All data report For Admin
func GetAllDataReports(PublicKey, MongoEnv, dbname, colname string, r *http.Request) string {
	req := new(Response)
	conn := SetConnection(MongoEnv, dbname)
	tokenlogin := r.Header.Get("Login")	
	if tokenlogin == "" {
		req.Status = false
		req.Message = "Header Login Not Found"
	} else {
		// Dekode token untuk mendapatkan
		_, err := DecodeGetReport(os.Getenv(PublicKey), tokenlogin)
		if err != nil {
			req.Status = false
			req.Message = "Tidak ada data  " + tokenlogin
		} else {
			// Langsung ambil data report
			datareport := GetAllReport(conn, colname)
			if datareport == nil {
				req.Status = false
				req.Message = "Data Report tidak ada"
			} else {
				req.Status = true
				req.Message = "Data Report berhasil diambil"
				req.Data = datareport
			}
		}
	}
	return ReturnStringStruct(req)
}

// get all report by Nik
func GCFGetAllReportID(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)

	var datareport Report
	err := json.NewDecoder(r.Body).Decode(&datareport)
	if err != nil {
		return err.Error()
	}

	report := GetAllReportID(mconn, collectionname, datareport)
	if report != (Report{}) {
		return GCFReturnStruct(CreateResponse(true, "Success: Get NIK Report", datareport))
	} else {
		return GCFReturnStruct(CreateResponse(false, "Failed to Get NIK Report", datareport))
	}
}


func GetOneDataReport(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	resp := new(Credential)
	reportdata := new(Report)
	resp.Status = false
	err := json.NewDecoder(r.Body).Decode(&reportdata)

	id := r.URL.Query().Get("_id")
	if id == "" {
		resp.Message = "Missing '_id' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		resp.Message = "Invalid '_id' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	reportdata.ID = ID

	// Menggunakan fungsi GetProdukFromID untuk mendapatkan data produk berdasarkan ID
	reportdata, err = GetReportFromID(mconn, collectionname, ID)
	if err != nil {
		resp.Message = err.Error()
		return GCFReturnStruct(resp)
	}

	resp.Status = true
	resp.Message = "Get Data Berhasil"
	resp.Data = []Report{*reportdata}

	return GCFReturnStruct(resp)
}

func GetOneDataReports(MONGOCONNSTRINGENV, dbname, collectionname string, r *http.Request) string {
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	resp := new(Credential)
	reportdata := new(Report)
	resp.Status = false

	err := json.NewDecoder(r.Body).Decode(&reportdata)

	nik := r.URL.Query().Get("nik")
	if nik == "" {
		resp.Message = "Missing 'nik' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	ID, err := strconv.Atoi(nik)
	if err != nil {
		resp.Message = "Invalid 'nik' parameter in the URL"
		return GCFReturnStruct(resp)
	}

	reportdata.Nik = ID

	// Menggunakan fungsi GetProdukFromID untuk mendapatkan data produk berdasarkan ID
	reportdata, err = GetReportFromIDs(mconn, collectionname, ID)
	if err != nil {
		resp.Message = err.Error()
		return GCFReturnStruct(resp)
	}

	resp.Status = true
	resp.Message = "Get Data Berhasil"
	resp.Data = []Report{*reportdata}

	return GCFReturnStruct(resp)
}


//Get One
// func GetOneReports(PublicKey, MongoEnv, dbname, colname string, r *http.Request) string {
// 	req := new(Response)
// 	conn := SetConnection(MongoEnv, dbname)
// 	tokenlogin := r.Header.Get("Login")
// 	if tokenlogin == "" {
// 		req.Status = false
// 		req.Message = "Header Login Not Found"
// 	} else {
// 		// Dekode token untuk mendapatkan
// 		_, err := DecodeGetReport(os.Getenv(PublicKey), tokenlogin)
// 		if err != nil {
// 			req.Status = false
// 			req.Message = "Tidak ada data  " + tokenlogin
// 		} else {
// 			// Langsung ambil data report
// 			datareport := GetOneReportData(conn, colname, Report{Nik: })
// 			if datareport == nil {
// 				req.Status = false
// 				req.Message = "Data Report tidak ada"
// 			} else {
// 				req.Status = true
// 				req.Message = "Data Report berhasil diambil"
// 				req.Data = datareport
// 			}
// 		}
// 	}
// 	return ReturnStringStruct(req)
// }




//Function Tanggapan

func GCFInsertTanggapan(publickey, MONGOCONNSTRINGENV, dbname, colladmin, colltanggapan string, r *http.Request) string {
	var response Credential
	response.Status = false
	mconn := SetConnection(MONGOCONNSTRINGENV, dbname)
	var admindata Admin
	gettoken := r.Header.Get("Login")
	if gettoken == "" {
		response.Message = "Missing Login in headers"
	} else {
		// Process the request with the "Login" token
		checktoken := watoken.DecodeGetId(os.Getenv(publickey), gettoken)
		admindata.Username = checktoken
		if checktoken == "" {
			response.Message = "Invalid token"
		} else {
			admin2 := FindAdmin(mconn, colladmin, admindata)
			if admin2.Role == "admin" {
				var datatanggapan Tanggapan
				err := json.NewDecoder(r.Body).Decode(&datatanggapan)
				if err != nil {
					response.Message = "Error parsing application/json: " + err.Error()
				} else {
					insertTanggapan(mconn, colltanggapan, Tanggapan{
						Nik:     		datatanggapan.Nik,
						Description: 	datatanggapan.Description,
						DateRespons: 	datatanggapan.DateRespons,
					})
					response.Status = true
					response.Message = "Berhasil Insert Tanggapan"
				}
			} else {
				response.Message = "Anda tidak bisa Insert Tanggapan karena bukan Admin"
			}
		}
	}
	return GCFReturnStruct(response)
}
