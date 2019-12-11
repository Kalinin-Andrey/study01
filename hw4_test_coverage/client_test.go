package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type UserParse struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}
type UsersParse struct {
	XMLName xml.Name    `xml:"root"`
	Users   []UserParse `xml:"row"`
}

type TestCase struct {
	ID          string
	AccessToken string
	Request     *SearchRequest
	Result      *SearchResponse
	IsError     bool
	Error       *SearchErrorResponse
}

var testCases []TestCase = []TestCase{
	TestCase{
		ID:          "success_empty_request",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: true,
		},
		IsError: false,
		Error:   nil,
	},
	TestCase{
		ID:          "success_next_page_false",
		AccessToken: "qwerty",
		Request: &SearchRequest{
			Query: "Guerrero",
			Limit: 10,
		},
		Result: &SearchResponse{
			Users: []User{
				User{
					Id:     12,
					Name:   "CruzGuerrero",
					Age:    36,
					About:  "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n",
					Gender: "male",
				},
			},
			NextPage: false,
		},
		IsError: false,
		Error:   nil,
	},
	TestCase{
		ID:          "error_broken_json",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "cant unpack result json: invalid character ']' after top-level value",
		},
	},
	TestCase{
		ID:          "error_timeout",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "timeout for limit=1&offset=0&order_by=0&order_field=&query=",
		},
	},
	TestCase{
		ID:          "error_unknown_error",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "unknown error Get",
		},
	},
	TestCase{
		ID:          "bad_access_token",
		AccessToken: "badtoken",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "Bad AccessToken",
		},
	},
	TestCase{
		ID:          "error_fatal_error",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "SearchServer fatal error",
		},
	},
	TestCase{
		ID:          "cant_unpack_error_json",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "cant unpack error json: invalid character 'c' looking for beginning of value",
		},
	},
	TestCase{
		ID:          "OrderFeld_invalid",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "OrderFeld  invalid",
		},
	},
	TestCase{
		ID:          "error_unknown_bad_request_error",
		AccessToken: "qwerty",
		Request:     &SearchRequest{},
		Result:      nil,
		IsError:     true,
		Error: &SearchErrorResponse{
			Error: "unknown bad request error: bad_request_error",
		},
	},
	TestCase{
		ID:          "limit_is_sub_zero",
		AccessToken: "qwerty",
		Request: &SearchRequest{
			Limit: -1,
		},
		Result:  nil,
		IsError: true,
		Error: &SearchErrorResponse{
			Error: "limit must be > 0",
		},
	},
	TestCase{
		ID:          "limit_is_up_to_25",
		AccessToken: "qwerty",
		Request: &SearchRequest{
			Limit: 30,
		},
		Result: &SearchResponse{
			Users: []User{
				User{
					Id:     0,
					Name:   "BoydWolf",
					Age:    22,
					About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
					Gender: "male",
				}, User{
					Id:     1,
					Name:   "HildaMayer",
					Age:    21,
					About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
					Gender: "female",
				}, User{
					Id:     2,
					Name:   "BrooksAguilar",
					Age:    25,
					About:  "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n",
					Gender: "male",
				}, User{
					Id:     3,
					Name:   "EverettDillard",
					Age:    27,
					About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
					Gender: "male",
				}, User{
					Id:     4,
					Name:   "OwenLynn",
					Age:    30,
					About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
					Gender: "male",
				}, User{
					Id:     5,
					Name:   "BeulahStark",
					Age:    30,
					About:  "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n",
					Gender: "female",
				}, User{
					Id:     6,
					Name:   "JenningsMays",
					Age:    39,
					About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
					Gender: "male",
				}, User{
					Id:     7,
					Name:   "LeannTravis",
					Age:    34,
					About:  "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n",
					Gender: "female",
				}, User{
					Id:     8,
					Name:   "GlennJordan",
					Age:    29,
					About:  "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n",
					Gender: "male",
				}, User{
					Id:     9,
					Name:   "RoseCarney",
					Age:    36,
					About:  "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n",
					Gender: "female",
				}, User{
					Id:     10,
					Name:   "HendersonMaxwell",
					Age:    30,
					About:  "Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n",
					Gender: "male",
				}, User{
					Id:     11,
					Name:   "GilmoreGuerra",
					Age:    32,
					About:  "Labore consectetur do sit et mollit non incididunt. Amet aute voluptate enim et sit Lorem elit. Fugiat proident ullamco ullamco sint pariatur deserunt eu nulla consectetur culpa eiusmod. Veniam irure et deserunt consectetur incididunt ad ipsum sint. Consectetur voluptate adipisicing aute fugiat aliquip culpa qui nisi ut ex esse ex. Sint et anim aliqua pariatur.\n",
					Gender: "male",
				}, User{
					Id:     12,
					Name:   "CruzGuerrero",
					Age:    36,
					About:  "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n",
					Gender: "male",
				}, User{
					Id:     13,
					Name:   "WhitleyDavidson",
					Age:    40,
					About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
					Gender: "male",
				}, User{
					Id:     14,
					Name:   "NicholsonNewman",
					Age:    23,
					About:  "Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n",
					Gender: "male",
				}, User{
					Id:     15,
					Name:   "AllisonValdez",
					Age:    21,
					About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
					Gender: "male",
				}, User{
					Id:     16,
					Name:   "AnnieOsborn",
					Age:    35,
					About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
					Gender: "female",
				}, User{
					Id:     17,
					Name:   "DillardMccoy",
					Age:    36,
					About:  "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n",
					Gender: "male",
				}, User{
					Id:     18,
					Name:   "TerrellHall",
					Age:    27,
					About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
					Gender: "male",
				}, User{
					Id:     19,
					Name:   "BellBauer",
					Age:    26,
					About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
					Gender: "male",
				}, User{
					Id:     20,
					Name:   "LoweryYork",
					Age:    27,
					About:  "Dolor enim sit id dolore enim sint nostrud deserunt. Occaecat minim enim veniam proident mollit Lorem irure ex. Adipisicing pariatur adipisicing aliqua amet proident velit. Magna commodo culpa sit id.\n",
					Gender: "male",
				}, User{
					Id:     21,
					Name:   "JohnsWhitney",
					Age:    26,
					About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
					Gender: "male",
				}, User{
					Id:     22,
					Name:   "BethWynn",
					Age:    31,
					About:  "Proident non nisi dolore id non. Aliquip ex anim cupidatat dolore amet veniam tempor non adipisicing. Aliqua adipisicing eu esse quis reprehenderit est irure cillum duis dolor ex. Laborum do aute commodo amet. Fugiat aute in excepteur ut aliqua sint fugiat do nostrud voluptate duis do deserunt. Elit esse ipsum duis ipsum.\n",
					Gender: "female",
				}, User{
					Id:     23,
					Name:   "GatesSpencer",
					Age:    21,
					About:  "Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n",
					Gender: "male",
				}, User{
					Id:     24,
					Name:   "GonzalezAnderson",
					Age:    33,
					About:  "Quis consequat incididunt in ex deserunt minim aliqua ea duis. Culpa nisi excepteur sint est fugiat cupidatat nulla magna do id dolore laboris. Aute cillum eiusmod do amet dolore labore commodo do pariatur sit id. Do irure eiusmod reprehenderit non in duis sunt ex. Labore commodo labore pariatur ex minim qui sit elit.\n",
					Gender: "male",
				},
			},
			NextPage: true,
		},
		IsError: false,
		Error:   nil,
	},
	TestCase{
		ID:          "offset_is_sub_zero",
		AccessToken: "qwerty",
		Request: &SearchRequest{
			Offset: -1,
		},
		Result:  nil,
		IsError: true,
		Error: &SearchErrorResponse{
			Error: "offset must be > 0",
		},
	},
}
var CurentTestCaseId string

func TestSearchServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchUsers))

	searchClient := SearchClient{
		AccessToken: "qwerty",
		URL:         ts.URL,
	}

	for caseNum, testCase := range testCases {
		searchClient.AccessToken = testCase.AccessToken
		CurentTestCaseId = testCase.ID
		response, err := searchClient.FindUsers(*testCase.Request)

		fmt.Printf("TestCase: #%d\t%v\n", caseNum, testCase.ID)
		// fmt.Println("response: ", response)
		// fmt.Println("err: ", err)

		if testCase.IsError == false && err != nil {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}

		if testCase.IsError == true && err == nil {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}

		if !reflect.DeepEqual(response, testCase.Result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, testCase.Result, response)
		}

		/*if testCase.IsError == true && !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}*/

		if testCase.IsError == true && !strings.Contains(err.Error(), testCase.Error.Error) {
			t.Errorf("[%d] wrong error, expected start with %#v, got %#v", caseNum, testCase.Error.Error, err.Error())
		}
	}
	/*if err != nil && !item.IsError {
		t.Errorf("[%d] unexpected error: %#v", caseNum, err)
	}
	if err == nil && item.IsError {
		t.Errorf("[%d] expected error, got nil", caseNum)
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
	}*/
	fmt.Printf("All done!\n\n")
	ts.Close()
}

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	var AccessTokenOrig string = "qwerty"
	var AccessToken string
	var Limit int
	var Offset int   // Можно учесть после сортировки
	var Query string // подстрока в 1 из полей

	var OrderField string
	// -1 по убыванию, 0 как встретилось, 1 по возрастанию
	var OrderBy int

	// fmt.Println("CurentTestCaseId=", CurentTestCaseId)

	AccessToken = r.Header.Get("AccessToken")
	Limit, err := strconv.Atoi(r.URL.Query().Get("limit"))

	if err != nil {
		panic(err)
	}
	Offset, err = strconv.Atoi(r.URL.Query().Get("offset"))

	if err != nil {
		panic(err)
	}
	Query = r.URL.Query().Get("query")
	OrderField = r.URL.Query().Get("order_field")
	OrderBy, err = strconv.Atoi(r.URL.Query().Get("order_by"))

	if err != nil {
		panic(err)
	}

	/*-----[	Validation	]-----*/
	if AccessTokenOrig != AccessToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if Limit < 0 {
		fmt.Println("Error: not valid Limit")
		return
	}

	if Offset < 0 {
		fmt.Println("Error: not valid Limit")
		return
	}

	switch OrderBy {
	case 0:
	case -1:
	case 1:
	default:
		fmt.Println("Error: not valid OrderBy")
		return
	}

	switch OrderField {
	case "":
		OrderField = "Name"
	case "Name":
	case "Id":
	case "Age":
	default:
		fmt.Println("Error: not valid OrderField")
		return
	}
	/*-----[	/Validation	]-----*/

	// Open our xmlFile
	xmlFile, err := os.Open("dataset.xml")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var usersParse UsersParse
	var users = make([]User, 0, 50)

	xml.Unmarshal(byteValue, &usersParse)
	//	make resulting type
	for _, item := range usersParse.Users {
		user := User{
			Id:     item.Id,
			Name:   item.FirstName + item.LastName,
			Age:    item.Age,
			About:  item.About,
			Gender: item.Gender,
		}
		if Query == "" || (Query != "" && strings.Contains(user.Name+";"+user.About, Query)) {
			users = append(users, user)
		}

	}

	switch OrderBy {
	case 0:
	case -1:
		switch OrderField {
		case "Name":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name > users[j].Name
			})
		case "Id":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Id > users[j].Id
			})
		case "Age":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age > users[j].Age
			})
		default:
			fmt.Println("err")
			return
		}
	case 1:
		switch OrderField {
		case "Name":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name < users[j].Name
			})
		case "Id":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Id < users[j].Id
			})
		case "Age":
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age < users[j].Age
			})
		default:
			fmt.Println("err")
			return
		}
	}
	var resultMax int

	if Offset+Limit < len(users) {
		resultMax = Offset + Limit
	} else {
		resultMax = len(users)
	}

	var catedUsers []User

	if Limit > 0 {
		catedUsers = users[Offset:resultMax]
	} else {
		catedUsers = users
	}

	/*fmt.Printf("\nId\t")
	fmt.Printf("Name\t\t")
	fmt.Printf("Age\n\n")

	for _, i := range catedUsers {
		fmt.Printf("%v\t", i.Id)
		fmt.Printf("%v\t", i.Name)
		fmt.Printf("%v\n", i.Age)
	}*/

	b, err := json.Marshal(catedUsers)

	if err != nil {
		panic(err)
	}

	if CurentTestCaseId == "error_broken_json" {
		b = b[1:]
	}

	if CurentTestCaseId == "error_timeout" {
		time.Sleep(2 * time.Second)
		return
	}

	if CurentTestCaseId == "error_fatal_error" {
		http.Error(w, "unknown error", http.StatusInternalServerError)
		return
	}

	if CurentTestCaseId == "error_unknown_error" {
		http.Error(w, "unknown error", http.StatusSeeOther)
		return
	}

	if CurentTestCaseId == "cant_unpack_error_json" {
		http.Error(w, "cant_unpack_error", http.StatusBadRequest)
		return
	}

	if CurentTestCaseId == "OrderFeld_invalid" {
		errResp := SearchErrorResponse{
			Error: "ErrorBadOrderField",
		}
		errRespJSON, err := json.Marshal(errResp)

		if err != nil {
			panic(err)
		}
		http.Error(w, string(errRespJSON), http.StatusBadRequest)
		return
	}

	if CurentTestCaseId == "error_unknown_bad_request_error" {
		errResp := SearchErrorResponse{
			Error: "bad_request_error",
		}
		errRespJSON, err := json.Marshal(errResp)

		if err != nil {
			panic(err)
		}
		http.Error(w, string(errRespJSON), http.StatusBadRequest)
		return
	}

	w.Write(b)

}
