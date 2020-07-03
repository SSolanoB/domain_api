package dbsetup

import (
	"fmt"
	"log"
  "strings"
  "strconv"

	"database/sql"
  "github.com/lib/pq"

  "context"
  "github.com/cockroachdb/cockroach-go/crdb"
  "./whoislocal"
  "./ssllabsapi"
  "./htmlreader"
)

type Server struct {
  Address string `json:"address"`
  Ssl_grade string `json:"ssl_grade"`
  Country string `json:"country"`
  Owner string `json:"owner"`
}

type Servers []Server

type Answer struct {
  Servers Servers `json:"servers"`
  Servers_changed bool `json:"servers_changed"`
  Ssl_grade string `json:"ssl_grade"`
  Previous_ssl_grade string `json:"previous_ssl_grade"`
  Logo string `json:"logo"`
  Title string `json:"title"`
  Is_down bool `json:"title"`
}

func ExecuteTransaction(resp ssllabsapi.Response) (r Answer, err error) {
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

    last_inquiry_id, err_lastin := checkLastInquiry(tx, url)

    if err_lastin != nil {
      return err_lastin
    }

    var previous_ssl_grade *string
    var err_check error
    if last_inquiry_id != nil {
      previous_ssl_grade, err_check = checkPreviousGrade(tx, last_inquiry_id)
      if err_check != nil {
        return err_check
      }
    }

    fmt.Println(last_inquiry_id)
    fmt.Println(previous_ssl_grade)

    if domain_id == nil {
      return fmt.Errorf("Not found? Found?")
    } else {
      fmt.Printf("Domain id is: %p\n", domain_id)
      if err := tx.QueryRow("INSERT INTO inquiries (domain_id, previous_ssl_grade, logo, title, is_down, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, now(), now()) RETURNING id", domain_id, previous_ssl_grade, logo, title, is_down).Scan(&inquiry_id); err != nil {
        return err
      }
      fmt.Printf("Inquiry id is: %p\n", inquiry_id)

      // I have to check if servers have changes

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
        var string_min_grade string
        for grade, value := range class {
          if min_grade != nil && *min_grade == value {
            string_min_grade = grade
            fmt.Println(string_min_grade)
            fmt.Println(grade)
          }
        }
        if _, err := tx.Exec("UPDATE inquiries SET ssl_grade = $1 WHERE id = $2", string_min_grade, inquiry_id); err != nil {
          return err
        }

        if last_inquiry_id != nil {
          servers_changed, err_serv := checkServersChanged(tx, last_inquiry_id, inquiry_id)
          
          if err_serv != nil {
            return err_serv
          }

          if *servers_changed == false {
            fmt.Println("ES FALSE, THEY DID NOT CHANGED")
          }
          if _, err := tx.Exec("UPDATE inquiries SET servers_changed = $1 WHERE id = $2", servers_changed, inquiry_id); err != nil {
            return err
          }
        }
      }
    }
  } else {
    fmt.Printf("Nor url provided")
  }

  return nil
}

func checkLastInquiry(tx *sql.Tx, url string) (last_inquiry_id *int, error error) {
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

  return last_inquiry_id, nil
}

func checkPreviousGrade(tx *sql.Tx, last_inquiry_id *int) (previous_ssl_grade *string, error error) {

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

  return previous_ssl_grade, nil
}

func checkServersChanged(tx *sql.Tx, last_inquiry_id *int, current_inquiry_id *int) (servers_changed *bool, error error) {
  p_true := true
  p_false := false

  rows, error := tx.Query("SELECT inquiry_id, address, ssl_grade, country, owner FROM servers WHERE inquiry_id = $1 OR inquiry_id = $2 ORDER BY address", last_inquiry_id, current_inquiry_id)
  defer rows.Close()

  switch {
  case error == sql.ErrNoRows:
    log.Println("no servers\n")
    return nil, nil
  case error != nil:
    log.Fatalf("query error: %v\n", error)
    return nil, nil
  default:
    log.Printf("Nothing in here")
  }

  var b_addresses, b_ssl_grades, b_countries, b_owners []string
  var c_addresses, c_ssl_grades, c_countries, c_owners []string

  for rows.Next() {
    var inquiry_id, address, ssl_grade, country, owner string

    if err := rows.Scan(&inquiry_id, &address, &ssl_grade, &country, &owner); err != nil {
      return nil, err
    }

    if inquiry_id == strconv.Itoa(*last_inquiry_id) {
      b_addresses = append(b_addresses, address)
      b_ssl_grades = append(b_ssl_grades, ssl_grade)
      b_countries = append(b_countries, country)
      b_owners = append(b_owners, owner)
    } else {
      c_addresses = append(c_addresses, address)
      c_ssl_grades = append(c_ssl_grades, ssl_grade)
      c_countries = append(c_countries, country)
      c_owners = append(c_owners, owner)
    }
  }

  if Identical(b_addresses, c_addresses) && Identical(b_ssl_grades, c_ssl_grades) && Identical(b_countries, c_countries) && Identical(b_owners, c_owners) {
    return &p_false, nil
  }

  return &p_true, nil
}

func Identical(str_1, str_2 []string) bool {
  if len(str_1) != len(str_2) {
    return false
  }
  for i, str := range str_1 {
    if str != str_2[i] {
      return false
    }
  }
  return true
}

func constructJson(inquiry_id int) (r Answer, error error) {
  db, err := sql.Open("postgres", 
    "postgresql://ssolanob@localhost:26257/development?ssl=true&sslmode=require&sslrootcert=certs/ca.crt&sslkey=certs/client.ssolanob.key&sslcert=certs/client.ssolanob.crt")

  if err != nil {
    log.Fatal("error connecting to the database: ", err)
  }
  defer db.Close()

  rows, error := db.Query("SELECT servers.address, servers.ssl_grade, servers.country, servers.owner, inquiries.servers_changed, inquiries.ssl_grade, inquiries.previous_ssl_grade, inquiries.logo, inquiries.title, inquiries.is_down FROM inquiries LEFT JOIN servers ON servers.inquiry_id = inquiries.id WHERE inquiries.id = $1", inquiry_id)
  defer rows.Close()

  switch {
  case error == sql.ErrNoRows:
    log.Println("no servers\n")
    return r, nil
  case error != nil:
    log.Fatalf("query error: %v\n", error)
    return r, nil
  default:
    log.Printf("Nothing in here")
  }

  notLast := rows.Next()

  for notLast {
    var address, ssl_grade, country, owner, servers_changed, min_ssl_grade, previous_ssl_grade, logo, title, is_down *string

    if err := rows.Scan(&address, &ssl_grade, &country, &owner, &servers_changed, &min_ssl_grade, &previous_ssl_grade, &logo, &title, &is_down); err != nil {
      return r, err
    } else {
      notLast = rows.Next()
      if notLast == false {
        if servers_changed != nil {
          sc_bool, err := strconv.ParseBool(*servers_changed)
          if err != nil {
            fmt.Println("Error with bool")
          } else {
            r.Servers_changed = sc_bool
          }  
        }
        
        if min_ssl_grade != nil {
          r.Ssl_grade = *min_ssl_grade  
        }
        
        if previous_ssl_grade != nil {
          r.Previous_ssl_grade = *previous_ssl_grade  
        }
        
        if logo != nil {
          r.Logo = *logo  
        }
        
        if title != nil {
          r.Title = *title
        }
        
        if is_down != nil {
          id_bool, err := strconv.ParseBool(*is_down)
          if err != nil {
            fmt.Println("Error with bool is down")
          } else {
            r.Is_down = id_bool
          }
        }
      }
      s := Server{}
      if address != nil {
        s.Address = *address
      }

      if ssl_grade != nil {
        s.Ssl_grade = *ssl_grade
      }

      if country != nil {
        s.Country = *country
      }

      if owner != nil {
        s.Owner = *owner
      }
      r.Servers = append(r.Servers, s)
    }
  }

  return r, nil

}

