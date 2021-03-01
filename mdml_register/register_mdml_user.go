package main

import (
	"log"
	"strconv"
	"strings"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"net/http"
	"os"
	"os/exec"
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
)

var AUTO_CREATE_IDS, err2 = strconv.ParseBool(os.Getenv("AUTO_CREATE_IDS"))
var ALLOW_TEST = os.Getenv("ALLOW_TEST")
var USE_BIS, err3 = strconv.ParseBool(os.Getenv("USE_BIS"))
var HOST = os.Getenv("HOSTNAME")
var MQTT_PSSWRD = os.Getenv("MDML_NODE_MQTT_USER")
var GRAFANA_PSSWRD = os.Getenv("MDML_GRAFANA_SECRET")
var BASIC_AUTH = base64.StdEncoding.EncodeToString([]byte("admin:" + GRAFANA_PSSWRD))

func registerUserResponse(w http.ResponseWriter, r *http.Request) {
	// Ignore invalid certificates
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Setting up response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization")

	// Respond to preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Only continue with POST requests
	if r.Method != "POST" {
		log.Printf(r.Method)
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("POST request received")
	
	// Get username and password
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
		return
	}
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	login_creds := strings.SplitN(string(payload), ":", 2)
	if len(login_creds) != 2 {
		http.Error(w, "authorization error", http.StatusUnauthorized)
		return
	}
	uname := login_creds[0]
	passwd := login_creds[1]
	
	log.Printf("Username and password received")
	
	// Get other data entries
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	dat := strings.SplitN(string(body), "&", 3)
	realname := strings.SplitN(string(dat[0]), "=", 2)[1]
	email := strings.SplitN(string(dat[1]), "=", 2)[1]
	experiment_id := strings.SplitN(string(dat[2]), "=", 2)[1]
	
	log.Printf("Other user data read")
	// Create MQTT user
	create_mqtt_userpass := exec.Command("mosquitto_passwd", "-b", "/etc/mosquitto/wordpassfile.txt", uname, passwd)
	err = create_mqtt_userpass.Run()
	if err != nil {
		log.Printf("MQTT: Error in userpass creation: %v \n", err)
		http.Error(w, "Error in MQTT user creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MQTT: User creation successful for: %v \n", uname)
	}


	// Create MQTT user's ACL entry
	create_mqtt_user_acl := exec.Command("/root/add_mqtt_acl.sh", uname, strings.ToUpper(experiment_id), ALLOW_TEST)
	err = create_mqtt_user_acl.Run()
	if err != nil {
		log.Printf("MQTT: Error in user ACL creation: %v \n", err)
		http.Error(w, "Error in MQTT user ACL creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MQTT: User ACL creation successful for: %v \n", uname)
	}


	// Create MinIO user
	create_minio_user := exec.Command("mc", "admin", "user", "add", "myminio", uname, passwd)
	err = create_minio_user.Run()
	if err != nil {
		log.Printf("MINIO: Error in user creation: %v \n", err)
		http.Error(w, "Error in MinIO user creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: User creation successful: %v \n", experiment_id)
	}


	// Create MinIO bucket
	create_minio_bucket := exec.Command("mc", "mb", "--ignore-existing", "myminio/mdml-"+strings.ToLower(experiment_id))
	err = create_minio_bucket.Run()
	if err != nil {
		log.Printf("MINIO: Error in bucket creation: %v \n", err)
		http.Error(w, "Error in MinIO bucket creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: Bucket creation successful: %v \n", experiment_id)
	}


	// Create MinIO bucket policy file
	create_policy_file := exec.Command("python",
		"/root/create_bucket_policy.py", experiment_id)
	err = create_policy_file.Run()
	if err != nil {
		log.Printf("MINIO: Error in policy file creation: %v \n", err)
		http.Error(w, "Error in MinIO policy file creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: Policy file creation successful: %v \n", experiment_id)
	}
	
	// Create policy with MinIO
	create_policy := exec.Command("mc", "admin", "policy", "add", "myminio", "readwrite_"+experiment_id, "/root/MinIO_policies/readwrite_"+experiment_id+".json")
	err = create_policy.Run()
	if err != nil {
		log.Printf("MINIO: Error in policy creation: %v \n", err)
		http.Error(w, "Error in MinIO policy creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: Policy creation successful: %v \n", experiment_id)
	}

	// Create MinIO group
	create_group := exec.Command("mc", "admin", "group", "add", "myminio", "readwrite_"+experiment_id, uname)
	err = create_group.Run()
	if err != nil {
		log.Printf("MINIO: Error in group creation: %v \n", err)
		http.Error(w, "Error in MinIO group creation. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: Group creation successful: %v \n", experiment_id)
	}

	// Attach policy to group
	attach_policy := exec.Command("mc", "admin", "policy", "set", "myminio", "readwrite_"+experiment_id, "group=readwrite_"+experiment_id)
	err = attach_policy.Run()
	if err != nil {
		log.Printf("MINIO: Error in attaching policy: %v \n", err)
		http.Error(w, "Error in MinIO attaching policy. Contact the MDML instance admin.", 500)
		return
	} else {
		log.Printf("MINIO: Attaching policy successful: %v \n", experiment_id)
	}
		
	// attach_user_policy := exec.Command("mc", "admin", "group", "add", "myminio", "readwrite_"+experiment_id, uname)
	// err = attach_user_policy.Run()
	// if err != nil {
	// 	log.Printf("MINIO: Error attaching user policy: %v \n", err)
	// 	http.Error(w, "Error in MinIO policy attachment. Contact the MDML instance admin.", 500)
	// 	return
	// } else {
	// 	log.Printf("MINIO: Policy attachment successful: %v \n", experiment_id)
	// }

	if USE_BIS {
		create_bis_bucket := exec.Command("s3cmd", "mb", "s3://mdml-"+strings.ToLower(experiment_id))
		err = create_bis_bucket.Run()
		if err != nil {
			log.Printf("BIS S3: Error creating bucket: %v \n", err)
			http.Error(w, "Error in BIS S3 bucket creation. Contact the MDML instance admin.", 500)
			return
		} else {
			log.Printf("BIS S3: Bucket creation successful: %v \n", "mdml-"+strings.ToLower(experiment_id))
		}
	}
	team_id := grafana_create_team(experiment_id)
	if team_id == -1 {
		http.Error(w, "Error in Grafana team creation. Contact the MDML instance admin.", 500)
		return
	}

	user_id := grafana_create_user(realname, email, uname, passwd)
	if user_id == -1 {
		http.Error(w, "Error in Grafana user creation. Contact the MDML instance admin.", 500)
		return
	}

	editor := grafana_user_role_editor(user_id)
	if !editor {
		http.Error(w, "Error in changing the user's role to editor. Contact the MDML instance admin.", 500)
		return
	}

	added := grafana_team_add_user(team_id, user_id)
	if !added {
		http.Error(w, "Error in adding user to Grafana team. Contact the MDML instance admin.", 500)
		return
	}

	dash_id := grafana_create_dashboard(experiment_id)
	if dash_id == -1 {
		http.Error(w, "Error creating the Grafana dashboard. Contact the MDML instance admin.", 500)
		return
	}

	permissions := grafana_add_dashboard_permissions(dash_id, team_id)
	if !permissions {
		http.Error(w, "Error adding permissions to the Grafana dashboard. Contact the MDML instance admin.", 500)
		return
	}

	// Sending message to NodeRED to create experiment ID if set_env.sh variables allow it
	if AUTO_CREATE_IDS {
		log.Printf("AUTO CREATING ID: %v \n", strings.ToLower(experiment_id))
		connOpts := mqtt.NewClientOptions().AddBroker("tcp://mdml_mosquitto_1:1883").SetUsername("nodered").SetPassword(MQTT_PSSWRD)
		client := mqtt.NewClient(connOpts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			http.Error(w, "Error1 creating experiment ID. Contact the MDML instance admin.", 500)
		}
		if token := client.Publish("ADMIN_MDML/EXPERIMENT", 2, false, experiment_id); token.Wait() && token.Error() != nil {
			http.Error(w, "Error2 creating experiment ID. Contact the MDML instance admin.", 500)
		}
		// http.Error(w, "Error sending auto create message.", 500)
		return
	}

}


func grafana_team_add_user(team_id int, user_id int) bool {
	mdml_url := "https://" + HOST + ":3000/api/teams/" + strconv.Itoa(team_id) + "/members"

	payload := strings.NewReader(`{"userId": `+ strconv.Itoa(user_id) + `}`)

	req, _ := http.NewRequest("POST", mdml_url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", HOST)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error in HTTP response for adding user to team.\n")
		return false
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		log.Printf("GRAFANA: User added to team successfully.\n")
		return true
	case 400:
		log.Printf("GRAFANA: User already exists on this team.\n")
		return false
	case 401:
		log.Printf("GRAFANA: Unathorized access adding user to team.\n")
		return false
	case 403:
		log.Printf("GRAFANA: Permission denied adding user to team.\n")
		return false
	case 404:
		log.Printf("GRAFANA: Team not found.\n")
		return false
	default:
		log.Printf("GRAFANA: Unknown status code when adding user to team.\n")
		return false
	}
}

func grafana_create_user(name string, email string, username string, password string) int {
	
	mdml_url := "https://" + HOST + ":3000/api/admin/users"

	v := url.Values{}
	v.Set("name", name)
	v.Add("email", email)
	v.Add("login", username)
	v.Add("password", password)
	payload := strings.NewReader(v.Encode())
	
	req, _ := http.NewRequest("POST", mdml_url, payload)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("GRAFANA: Error in user creation: %v \n", err)
		return -1
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode == 200 {
		// Expected response if successful
		type Add_User_Response struct {
			ID int			`json:"id"`
			Message string	`json:"message"`
		}
		data := Add_User_Response{}
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Printf("GRAFANA: Error in user creation: %v \n", err)
		}
		log.Printf("GRAFANA: User creation successful: %v \n", username)
		return data.ID
	} else {
		log.Printf("GRAFANA: Error in user creation: %v \n", string(body))
		return -1
	}
}

func grafana_get_team_id(experiment_id string) int {
	
	mdml_url := "https://" + HOST + ":3000/api/teams/search?name=" + experiment_id

	req, _ := http.NewRequest("GET", mdml_url, nil)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", HOST)
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error in HTTP response for getting team ID.")
		return -1
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Printf(string(body))

	switch res.StatusCode {
	case 200:
		// Expected response if successful
		type teams struct {
			ID int 				`json:"id"`
			OrgID int			`json:"orgId"`
			Name string			`json:"name"`
			Email string 		`json:"email"`
			AvatarURL string	`json:"avatarUrl"`
			MemberCount int		`json:"memberCount"`
			Permission int		`json:"permission"`
		}
		type TeamID_Response struct {
			TotalCount int		`json:"totalCount"`
			Teams [1]teams 		`json:"teams"`
			Page int			`json:"page"`
			PerPage int			`json:"perPage"`
		}
		data := TeamID_Response{}
		// string to object. return -1 if errs
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Printf("GRAFANA: Error in team creation: %v \n", err)
			return -1
		}
		// return Team ID
		log.Printf("GRAFANA: Team ID found for: %v \n", experiment_id)
		return data.Teams[0].ID

	case 401:
		log.Printf("GRAFANA: Unathorized access when getting team ID.\n")
		return -1
	case 403:
		log.Printf("GRAFANA: Permission denied when getting team ID.\n")
		return -1
	case 404:
		log.Printf("GRAFANA: Team name not found: %v \n", experiment_id)
		return -1
	default:
		log.Printf("GRAFANA: Unknown status code when getting team ID.\n")
		return -1
	}
}

func grafana_user_role_editor(user_id int) bool {
	mdml_url := "https://" + HOST + ":3000/api/org/users/" + strconv.Itoa(user_id)

	payload := strings.NewReader(`{"role": "Editor"}`)

	req, _ := http.NewRequest("PATCH", mdml_url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("cache-control", "no-cache")
	
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("GRAFANA: Error in changing user role: %v \n", err)
		return false
	}

	if res.StatusCode == 200 {
		log.Printf("GRAFANA: User role changed to editor.")
		return true
	} else {
		log.Printf("GRAFANA: Error in changing user role.")
		return false
	}
}

func grafana_create_dashboard(experiment_id string) int {
	mdml_url := "https://" + HOST + ":3000/api/dashboards/db"

	payload := strings.NewReader(`{
		"dashboard": {
		  "id": null,
		  "uid": null,
		  "title": "` + experiment_id + ` Dashboard",
		  "timezone": "browser",
		  "refresh": "1s,5s,10s,30s"
		},
		"overwrite": false
	}`)

	req, _ := http.NewRequest("POST", mdml_url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error in creating dashboard: %v \n", err)
		return -1
	}
	
	// if err != nil {
	// 	log.Printf("GRAFANA: Error in HTTP response for getting team ID.")
	// 	return -1
	// }
	
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)


	if res.StatusCode == 200 {
		// Expected response if successful
		type dashboard_response struct {
			ID int 				`json:"id"`
			UID string			`json:"uid"`
			URL string			`json:"url"`
			Status string 		`json:"status"`
			Version int			`json:"version"`
		}
		data := dashboard_response{}
		// string to object. return -1 if errs
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Printf("GRAFANA: Error in parsing dashboard creation response: %v \n", body)
			return -1
		}

		log.Printf("GRAFANA: Dashboard created.")
		return data.ID
	} else if res.StatusCode == 412 {
		log.Printf("GRAFANA: Dashboard already exists. Need to find out its ID")
		
		mdml_url := "https://" + HOST + ":3000/api/search?query=" + experiment_id + "%20Dashboard"

		// payload := strings.NewReader(`{
		// 	"dashboardIds": ["` + experiment_id + ` Dashboard"]
		// }`)

		req, _ := http.NewRequest("GET", mdml_url, nil)
		
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
		
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("GRAFANA: Error searching for dashboard: %v \n", err)
			return -1
		}
		
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		log.Printf("%v", string(body))

		// Expected response if successful
		type dashboard struct {
			ID int 				`json:"id"`
			UID string			`json:"uid"`
			Title string		`json:"title"`
			URI string			`json:"uri"`
			URL string			`json:"url"`
			Slug string			`json:"slug"`
			Type string			`json:"type"`
			Tags []string		`json:"tags"`
			isStarred bool		`json:"isStarred"`
		}
		var dashdata []dashboard

		err = json.Unmarshal([]byte(body), &dashdata)
		if err != nil {
			log.Printf("GRAFANA: Error parsing dashboard search response: %v \n", err)
			return -1
		}
		log.Printf("%v", dashdata[0].ID)
		return dashdata[0].ID
	} else {
		log.Printf("GRAFANA: Error in creating dashboard.")
		return -1
	}

}

func grafana_add_dashboard_permissions(dashboard_id int, team_id int) bool {
	mdml_url := "https://" + HOST + ":3000/api/dashboards/id/" + strconv.Itoa(dashboard_id) + "/permissions"
	log.Printf("%v", mdml_url)
	payload := strings.NewReader(`{
		"items": [
			{
				"teamId": ` + strconv.Itoa(team_id) + `,
				"permission": 2 
			}
		]
	}`)

	req, err := http.NewRequest("POST", mdml_url, payload)
	if err!=nil {
		log.Printf("GRAFANA: Error creating dashboard permissions request: %v \n", err)
		return false
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error adding dashboard permissions: %v \n", err)
		return false
	}
	
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode == 200 {
		// // Expected response if successful
		// type perm struct {
		// 	Message int 		`json:"message"`
		// }
		// data := perm{}
		// // string to object. return -1 if errs
		// err := json.Unmarshal([]byte(body), &data)
		// if err != nil {
		// 	log.Printf("GRAFANA: Error parsing permission response: %v \n", err)
		// 	return false
		// }
		// // return Team ID
		log.Printf("GRAFANA: Dashboard permissions set.")
		return true
	} else {
		log.Printf(string(body))
		log.Printf("GRAFANA: Error %v when setting dashboard permissions.\n", res.StatusCode)
		return false
	}
}

func grafana_create_team(experiment_id string) int {
	
	log.Printf("HOST: %v \n", HOST)
	mdml_url := "https://" + HOST + ":3000/api/teams/"
	params := "name="
	params += experiment_id
	payload := strings.NewReader(params)
	
	req, _ := http.NewRequest("POST", mdml_url, payload)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", HOST)
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error in team creation: %v \n", err)
		return -1
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	switch res.StatusCode {
	case 200:
		// Expected response if successful
		type Grafana_Response struct {
			Message string		`json:"message"`
			TeamID int 			`json:"teamId"`
		}
		data := Grafana_Response{}
		// String to object
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Printf("GRAFANA: Error in team creation: %v \n", err)
			return -1
		}
		log.Printf("GRAFANA: Team creation successful: %v \n", experiment_id)
		return data.TeamID
	case 401:
		log.Printf("GRAFANA: Unathorized access when creating team.\n")
		return -1
	case 403:
		log.Printf("GRAFANA: Permission denied when creating team.\n")
		return -1
	case 409:
		log.Printf("GRAFANA: Team already exists. Getting team ID for: %v \n", experiment_id)
		team_id := grafana_get_team_id(experiment_id)
		return team_id
	default:
		log.Printf("GRAFANA: Unknown status code in team creation.\n")
		return -1
	}
}

func getUsers(w http.ResponseWriter, r *http.Request){
	// Ignore invalid certificates
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Setting up response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization")
	
	req, _ := http.NewRequest("GET", "https://" + HOST + "/grafana/api/org/users", nil)
	
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic " + BASIC_AUTH)
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", HOST)
	
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GRAFANA: Error in team creation: %v \n", err)
		return
	}
	
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	
	switch res.StatusCode {
	case 200:
		http.Error(w, string(body), 200)
	default:
		log.Printf("GRAFANA: Could not get users.\n")
		return
	}
}

func main() {
	http.HandleFunc("/", registerUserResponse)
	http.HandleFunc("/users", getUsers)
	// http.ListenAndServe(":8184", nil)
	http.ListenAndServeTLS(":8184", "/etc/ssl/certs/nginx-selfsigned.crt", "/etc/ssl/nginx-selfsigned.key", nil)
}
