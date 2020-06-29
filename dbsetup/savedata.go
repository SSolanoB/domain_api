package dbsetup

import (
	"fmt"
	"log"

	"database/sql"
  _ "github.com/lib/pq"

  "context"
  "github.com/cockroachdb/cockroach-go/crdb"
  "./whoislocal"
)

func ExecuteTransaction(resp Response) error {
  fmt.Println(resp)
  //fmt.Printf("Body is: %T\n", resp)

  db, err := sql.Open("postgres", 
    "postgresql://ssolanob@localhost:26257/development?ssl=true&sslmode=require&sslrootcert=certs/ca.crt&sslkey=certs/client.ssolanob.key&sslcert=certs/client.ssolanob.crt")

  if err != nil {
    log.Fatal("error connecting to the database: ", err)
  }
  defer db.Close()

  err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
      return SaveData(tx, resp)
  })
  if err == nil {
      fmt.Println("Success")
  } else {
      log.Fatal("error: ", err)
  }
  return err
}

func SaveData(tx *sql.Tx, resp Response) error {
    
  url := resp.Host

  if url != "" {
    var domain_id *int
    var inquiry_id *int

    if err := tx.QueryRow("SELECT id FROM domains WHERE url = $1", url).Scan(&domain_id); err != nil {
      if err_2 := tx.QueryRow("INSERT INTO domains (url, created_at, updated_at) VALUES ($1, now(), now()) RETURNING id", url).Scan(&domain_id); err_2 != nil {
        return err_2
      }
    }

    if domain_id == nil {
      return fmt.Errorf("Not found? Found?")
    } else {
      fmt.Printf("Domain id is: %p\n", domain_id)
      if err := tx.QueryRow("INSERT INTO inquiries (domain_id, created_at, updated_at) VALUES ($1, now(), now()) RETURNING id", domain_id).Scan(&inquiry_id); err != nil {
        return err
      }
      fmt.Printf("Inquiry id is: %p\n", inquiry_id)

      // I have to check the logo, title, and if the web is down.
      // I have to check if servers have changes, the min ssl grade, and the previous ssl grade if available.

      servers := resp.Endpoints
      
      class := map[string]int{
        "T": 0,
        "M": 1,
        "F": 2,
        "E": 3,
        "D": 4,
        "C": 5,
        "B": 6,
        "A-": 7,
        "A": 8,
        "A+": 9,
      }

      var min_grade *int

      if len(servers) > 0 {
        for i, server := range servers {
          if server.StatusMessage != "Ready" {
            continue;
          }
          var ssl_grade *string
          country, owner, err := whoislocal.AskforIp(server.IpAddress)
          if err == nil {
            if err := tx.QueryRow("INSERT INTO servers (inquiry_id, address, ssl_grade, created_at, updated_at) VALUES ($1, $2, $3, now(), now()) RETURNING ssl_grade", inquiry_id, server.IpAddress, server.Grade).Scan(&ssl_grade); err != nil {
              return err
            }
          } else {
            return err
          }
          
          fmt.Println(i)
          fmt.Println(ssl_grade)
          temp_ssl := *ssl_grade
          temp_grade := class[temp_ssl]

          if min_grade == nil {
            min_grade = &temp_grade
          } else {
            if temp_grade < *min_grade {
              min_grade = &temp_grade
            }
          }

          fmt.Println(temp_grade)
        }
        fmt.Println(*min_grade)
        var string_min_grade string
        for grade, value := range class {
          if *min_grade == value {
            string_min_grade = grade
            fmt.Println(string_min_grade)
            fmt.Println(grade)
          }
        }
        if _, err := tx.Exec("UPDATE inquiries SET ssl_grade = $1 WHERE id = $2", string_min_grade, inquiry_id); err != nil {
          return err
        }
      }
    }
  } else {
    fmt.Printf("Nor url provided")
  }

  // Perform the transfer.
  /*if _, err := tx.Exec(
    "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from); err != nil {
    return err
  }
  if _, err := tx.Exec(
    "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to); err != nil {
    return err
  }*/
  return nil
}