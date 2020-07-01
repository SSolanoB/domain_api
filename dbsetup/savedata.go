package dbsetup

import (
	"fmt"
	"log"

	"database/sql"
  _ "github.com/lib/pq"

  "context"
  "github.com/cockroachdb/cockroach-go/crdb"
  "./whoislocal"
  "./ssllabsapi"
  "./htmlreader"
)

func ExecuteTransaction(resp ssllabsapi.Response) error {
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

func SaveData(tx *sql.Tx, resp ssllabsapi.Response) error {
    
  url := resp.Host

  if url != "" {
    var domain_id *int
    var inquiry_id *int

    if err := tx.QueryRow("SELECT id FROM domains WHERE url = $1", url).Scan(&domain_id); err != nil {
      if err_2 := tx.QueryRow("INSERT INTO domains (url, created_at, updated_at) VALUES ($1, now(), now()) RETURNING id", url).Scan(&domain_id); err_2 != nil {
        return err_2
      }
    }

    titles, images, links, is_down := htmlreader.RequestHeaderInfo(url)
    fmt.Println("Inside SaveData \n")
    fmt.Println(titles)
    fmt.Println(images)
    fmt.Println(links)

    var title string
    var logo string

    if titles != nil {
      title = titles[0]
    }

    if images != nil && links != nil {
      logo = links[0]
    } else if links != nil {
      logo = links[0]
    } else if images != nil {
      logo = images[0]
    }

    if domain_id == nil {
      return fmt.Errorf("Not found? Found?")
    } else {
      fmt.Printf("Domain id is: %p\n", domain_id)
      if err := tx.QueryRow("INSERT INTO inquiries (domain_id, logo, title, is_down, created_at, updated_at) VALUES ($1, $2, $3, $4, now(), now()) RETURNING id", domain_id, logo, title, is_down).Scan(&inquiry_id); err != nil {
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
            if err := tx.QueryRow("INSERT INTO servers (inquiry_id, address, ssl_grade, country, owner, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING ssl_grade", inquiry_id, server.IpAddress, server.Grade, country, owner).Scan(&ssl_grade); err != nil {
              return err
            }
          } else {
            if err := tx.QueryRow("INSERT INTO servers (inquiry_id, address, ssl_grade, created_at, updated_at) VALUES ($1, $2, $3, now(), now()) RETURNING ssl_grade", inquiry_id, server.IpAddress, server.Grade).Scan(&ssl_grade); err != nil {
              return err
            }
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

  return nil
}

func checkPreviousServers(tx *sql.Tx, url string) (previous_ssl_grade *string, error error) {
  
  trimmed_url := strings.Trim(url, "http://")
  trimmed_url = strings.Trim(trimmed_url, "https://")

  rows, error := tx.Query("SELECT id FROM domains WHERE url LIKE $1", "%" + trimmed_url + "%")

  if error != nil {
    return
  }
  defer rows.Close()

  var dom_ids []int

  for rows.Next() {
    var id int
    if error = rows.Scan(&id); error != nil {
      return
    }
    dom_ids = append(dom_ids, id)
  }

  fmt.Println("CHECK PREVIOUS DOMAINS 2")
  fmt.Println(dom_ids)

  var last_inquiry_id *int
  error = tx.QueryRow("SELECT inquiries.id FROM inquiries INNER JOIN domains ON domains.id = inquiries.domain_id AND domains.id = ANY ($1) WHERE inquiries.created_at < (now() - INTERVAL '1 HOUR') ORDER BY inquiries.created_at DESC LIMIT 1", pq.Array(dom_ids)).Scan(&last_inquiry_id)

  switch {
  case error == sql.ErrNoRows:
    log.Printf("no inquiries found with domain id in %d\n", dom_ids)
    return nil, nil
  case error != nil:
    log.Fatalf("query error: %v\n", error)
    return 
  default:
    log.Printf("Nothing in here")
  }
  
  fmt.Printf("Inquiry id is: %+v\n", *last_inquiry_id)

  error = tx.QueryRow("SELECT ssl_grade FROM inquiries WHERE id = $1", last_inquiry_id).Scan(&previous_ssl_grade)

  switch {
  case error == sql.ErrNoRows:
    log.Printf("no ssl_grade found with inquiry id %d\n", last_inquiry_id)
    return nil, nil
  case error != nil:
    log.Fatalf("query error: %v\n", error)
    return 
  default:
    log.Printf("Nothing in here")
  }


  /*var inq_ids []int

  for rows_in.Next() {
    var id int
    if err := rows_in.Scan(&id); err != nil {
      return err
    }
    inq_ids = append(inq_ids, id)
  }

  fmt.Println("CHECK PREVIOUS INQUIRIES")
  fmt.Println(inq_ids)*/


  return previous_ssl_grade, nil
}