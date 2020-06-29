package dbsetup

import (
	"log"

	"database/sql"
  _ "github.com/lib/pq"
)


func SetupDb() {
  // Connect to bank DB
  db, err := sql.Open("postgres", 
    "postgresql://ssolanob@localhost:26257/development?ssl=true&sslmode=require&sslrootcert=certs/ca.crt&sslkey=certs/client.ssolanob.key&sslcert=certs/client.ssolanob.crt")

  if err != nil {
    log.Fatal("error connecting to the database: ", err)
  }
  defer db.Close()

  if _, err := db.Exec(
    "CREATE TABLE IF NOT EXISTS domains (id SERIAL NOT NULL PRIMARY KEY, url STRING NOT NULL UNIQUE, created_at TIMESTAMP, updated_at TIMESTAMP)"); err != nil {
    log.Fatal(err)
  }

  if _, err := db.Exec(
    "CREATE TABLE IF NOT EXISTS inquiries (id SERIAL NOT NULL PRIMARY KEY, domain_id INT NOT NULL REFERENCES domains (id) ON DELETE CASCADE, servers_changed BOOL, ssl_grade STRING, previous_ssl_grade STRING, logo STRING, title STRING, is_down BOOL, created_at TIMESTAMP, updated_at TIMESTAMP, INDEX (domain_id))"); err != nil {
    log.Fatal(err)
  }
  
  if _, err := db.Exec(
    "CREATE TABLE IF NOT EXISTS servers (id SERIAL NOT NULL PRIMARY KEY, inquiry_id INT NOT NULL REFERENCES inquiries (id) ON DELETE CASCADE, address STRING, ssl_grade STRING, country STRING, owner STRING, created_at TIMESTAMP, updated_at TIMESTAMP, INDEX (inquiry_id))"); err != nil {
    log.Fatal(err)
  }
}