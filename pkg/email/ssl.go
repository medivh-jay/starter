package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

func dial() (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", email.host, email.port), nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	return smtp.NewClient(conn, email.host)
}

func sendBySsl(sender *Object) {
	c, err := dial()
	if !catch(err) {
		defer c.Close()
	}

	if ok, _ := c.Extension("AUTH"); ok {
		if err = c.Auth(auth); !catch(err) {
			return
		}
	}

	err = c.Mail(email.username)
	if !catch(err) {
		return
	}

	for _, addr := range sender.To {
		err = c.Rcpt(addr)
		if !catch(err) {
			return
		}
	}

	w, err := c.Data()
	if !catch(err) {
		return
	}

	_, err = w.Write(sender.convertToBody())
	if !catch(err) {
		return
	}

	err = w.Close()
	if !catch(err) {
		return
	}

	err = c.Quit()
	if err != nil {
		log.Println(err)
		return
	}
}

func catch(err error) bool {
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
