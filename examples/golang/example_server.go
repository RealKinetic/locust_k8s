// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package main

import (
	"fmt"
	"log"
	"net/http"
)

func index_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Index Request")
	fmt.Fprintf(w, "Our example home page.")
}

func login_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login Request")
	fmt.Fprintf(w, "Login.")
}

func profile_handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Profile Request")
	fmt.Fprintf(w, "Profile.")
}

func main() {
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/login", login_handler)
	http.HandleFunc("/profile", profile_handler)
	log.Printf("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
