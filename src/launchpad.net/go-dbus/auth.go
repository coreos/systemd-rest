package dbus

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"net"
	"strconv"
)

type Authenticator interface {
	Mechanism() []byte
	InitialResponse() []byte
	ProcessData(challenge []byte) (response []byte, err error)
}

type AuthExternal struct {
}

func (p *AuthExternal) Mechanism() []byte {
	return []byte("EXTERNAL")
}

func (p *AuthExternal) InitialResponse() []byte {
	uid := []byte(strconv.Itoa(os.Geteuid()))
	uidHex := make([]byte, hex.EncodedLen(len(uid)))
	hex.Encode(uidHex, uid)
	return uidHex
}

func (p *AuthExternal) ProcessData([]byte) ([]byte, error) {
	return nil, errors.New("Unexpected Response")
}

type AuthDbusCookieSha1 struct {
}

func (p *AuthDbusCookieSha1) Mechanism() []byte {
	return []byte("DBUS_COOKIE_SHA1")
}

func (p *AuthDbusCookieSha1) InitialResponse() []byte {
	user := []byte(os.Getenv("USER"))
	userHex := make([]byte, hex.EncodedLen(len(user)))
	hex.Encode(userHex, user)
	return userHex
}

func (p *AuthDbusCookieSha1) ProcessData(mesg []byte) ([]byte, error) {
	decodedLen, err := hex.Decode(mesg, mesg)
	if err != nil {
		return nil, err
	}
	mesgTokens := bytes.SplitN(mesg[:decodedLen], []byte(" "), 3)

	file, err := os.Open(os.Getenv("HOME") + "/.dbus-keyrings/" + string(mesgTokens[0]))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileStream := bufio.NewReader(file)

	var cookie []byte
	for {
		line, _, err := fileStream.ReadLine()
		if err == io.EOF {
			return nil, errors.New("SHA1 Cookie not found")
		} else if err != nil {
			return nil, err
		}
		cookieTokens := bytes.SplitN(line, []byte(" "), 3)
		if bytes.Compare(cookieTokens[0], mesgTokens[1]) == 0 {
			cookie = cookieTokens[2]
			break
		}
	}

	challenge := make([]byte, len(mesgTokens[2]))
	if _, err = rand.Read(challenge); err != nil {
		return nil, err
	}

	for temp := challenge; ; {
		if index := bytes.IndexAny(temp, " \t"); index == -1 {
			break
		} else if _, err := rand.Read(temp[index : index+1]); err != nil {
			return nil, err
		} else {
			temp = temp[index:]
		}
	}

	hash := sha1.New()
	if _, err := hash.Write(bytes.Join([][]byte{mesgTokens[2], challenge, cookie}, []byte(":"))); err != nil {
		return nil, err
	}

	resp := bytes.Join([][]byte{challenge, []byte(hex.EncodeToString(hash.Sum(nil)))}, []byte(" "))
	respHex := make([]byte, hex.EncodedLen(len(resp)))
	hex.Encode(respHex, resp)
	return respHex, nil
}

func authenticate(conn net.Conn, authenticators []Authenticator) error {
	// If no authenticators are provided, try them all
	if authenticators == nil {
		authenticators = []Authenticator{
			new(AuthExternal),
			new(AuthDbusCookieSha1)}
	}

	// The authentication process starts by writing a nul byte
	if _, err := conn.Write([]byte{0}); err != nil {
		return err
	}

	inStream := bufio.NewReader(conn)
	send := func(command ...[]byte) ([][]byte, error) {
		msg := bytes.Join(command, []byte(" "))
		_, err := conn.Write(append(msg, []byte("\r\n")...))
		if err != nil {
			return nil, err
		}
		line, isPrefix, err := inStream.ReadLine()
		if err != nil {
			return nil, err
		}
		if isPrefix {
			return nil, errors.New("Received line is too long")
		}
		return bytes.Split(line, []byte(" ")), err
	}
	success := false
	for _, auth := range authenticators {
		reply, err := send([]byte("AUTH"), auth.Mechanism(), auth.InitialResponse())
		StatementLoop:
		for {
			if err != nil {
				return err
			}
			if len(reply) < 1 {
				return errors.New("No response command from server")
			}
			switch string(reply[0]) {
			case "OK":
				success = true
				break StatementLoop
			case "REJECTED":
				// XXX: should note the list of
				// supported mechanisms
				break StatementLoop
			case "ERROR":
				return errors.New("Received error from server: " + string(bytes.Join(reply, []byte(" "))))
			case "DATA":
				var response []byte
				response, err = auth.ProcessData(reply[1])
				if err == nil {
					reply, err = send([]byte("DATA"), response)
				} else {
					// Cancel so we can move on to
					// the next mechanism.
					reply, err = send([]byte("CANCEL"))
				}
			default:
				return errors.New("Unknown response from server: " + string(bytes.Join(reply, []byte(" "))))
			}
		}
		if success {
			break
		}
	}
	if !success {
		return errors.New("Could not authenticate with any mechanism")
	}
	// XXX: UNIX FD negotiation would go here.
	if _, err := conn.Write([]byte("BEGIN\r\n")); err != nil {
		return err
	}
	return nil
}
