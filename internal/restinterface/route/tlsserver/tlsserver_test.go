/*******************************************************************************
 * Copyright 2021 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

package tlsserver

import (
	"os"
	"testing"
	"time"
)

const (
	fakeCertsPath     = "fakecerts"
	unexpectedSuccess = "unexpected success"
	unexpectedFail    = "unexpected fail"
)

const fakeCASert = `
-----BEGIN CERTIFICATE-----
MIIDjzCCAnegAwIBAgIUR/W/1J1nRNwoWY4Bt+a0UNOLYrMwDQYJKoZIhvcNAQEL
BQAwVzELMAkGA1UEBhMCS1IxDjAMBgNVBAgMBVNlb3VsMRwwGgYDVQQKDBNTYW1z
dW5nIEVsZWN0cm9uaWNzMRowGAYDVQQDDBFIb21lIEVkZ2UgQ0EgUm9vdDAeFw0y
MjA0MjAxNDU3NDVaFw0zMjA0MTcxNDU3NDVaMFcxCzAJBgNVBAYTAktSMQ4wDAYD
VQQIDAVTZW91bDEcMBoGA1UECgwTU2Ftc3VuZyBFbGVjdHJvbmljczEaMBgGA1UE
AwwRSG9tZSBFZGdlIENBIFJvb3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDMnisxz1qCvaTcvuIQh2Kgfb2bBNBvtm6t0Ls1IwtKp77J6Mt1xAo2dJnv
7glkVZulIQ4aCJ29dnaQcZAHWY0h79F9JF6TeJl1qKAGfXXxx14La6HN1t6kBlNE
76Ix3sEpko9n30sE6LmM7NxZAZ2A1Hk+m88ETl7W04PX7gG3ssl7JQtCfDEkrbgO
OtSWHL0JwAINdjBcZzRjtxIL24E4x1kl9nAdMF7GMWfIh+Wuhw3C4go2dFKx8jJq
a5L9oedFwTsPrVM4IX8gtQ3hrgyI+VQO2s0z+2hAmMNvjWHMIdoivDASlUYtK6y1
fj3l9Y5xnR+4xx4WfN7idC7Y9/HnAgMBAAGjUzBRMB0GA1UdDgQWBBSkztsQFzac
kFrXpqAw1hjabLfqbjAfBgNVHSMEGDAWgBSkztsQFzackFrXpqAw1hjabLfqbjAP
BgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQDL3ibEbFh4enlNcqJI
hiukqE/94n+rSBhYqd0ol5X+3NU67gCKgu0zbvbFZ7oR1Oh+hAwiymWt7w8AhhBG
ZNhVwRdsqX+QQm+TQZUIo9hXtCnMgNV+dSmk3bwVo5epdKmqoHztLng4o70AhXSK
XPth+yjNlxlfeFSLP0lBfhF821XYfhwL5LAgOBj0Pw7s7yWEeGfAVrYIw0QFx/fb
bjys6MZ4O6yiTfQtFwVyxYjnihsBbVeSVs/qx9Z1lafggsauHo76/McgEvVaZtPy
XINE2U0o4WTLMcl7RVzftmPCYQks5ogh/eyVcTiJS4u3Xor9XGPhvDop7FT4vl7B
GDnu
-----END CERTIFICATE-----`

const fakeHENSert = `
-----BEGIN CERTIFICATE-----
MIIDWDCCAkCgAwIBAgIUVkHr3nlXL0+aiQMsWCEi4hZZSRcwDQYJKoZIhvcNAQEL
BQAwVzELMAkGA1UEBhMCS1IxDjAMBgNVBAgMBVNlb3VsMRwwGgYDVQQKDBNTYW1z
dW5nIEVsZWN0cm9uaWNzMRowGAYDVQQDDBFIb21lIEVkZ2UgQ0EgUm9vdDAeFw0y
MjA0MjAxNDU4MTJaFw0yMzA0MjAxNDU4MTJaMGAxCzAJBgNVBAYTAktSMQ4wDAYD
VQQIDAVTZW91bDEcMBoGA1UECgwTU2Ftc3VuZyBFbGVjdHJvbmljczEjMCEGA1UE
AwwaSG9tZSBFZGdlIE5vZGUgQ2VydGlmaWNhdGUwggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQDAjuVJmWNsZJ4J++F+mEyKgLR0sQkJnFeWVD5l6cmj7r5C
R0n7CIDKI1lsvWmwLT7f5nOlomrD06/rg0srTpHEBdtb1uIa5C/IX6jXtjMxphfS
ZFEpx9+Kg/VQie9AZQVic2BHsGuNqTqqcaspTjPl1ttl4y9/sjurYrJ7BDUkfAyb
VO0j+fDdlSvd8wq1SweFaU39uu8GIqQ+IbQTVe5WiCkfEljkDKKnEqO0cieyoGMA
7DYEl31gC8WIhKhGZPnVDfHEJFINMSM8tKTehmWV5NWj9pRmfyffiyRpeFUGMNLq
tgJ1o8NEemyDhzDUlpqZ0OlhsgiOCz7zppjfNFC7AgMBAAGjEzARMA8GA1UdEQQI
MAaHBAoAAgQwDQYJKoZIhvcNAQELBQADggEBAGOpTa5WVoJyEQEKLey6lJjml2yi
Ga5HvevLqMjBLEj955VMFZjMF+uHaBv+//U45XJaZD8tUk130PO1IyE86XRYCatk
mQEtvpYjnjIUgE3cWKoqyjkdk73Hg+LRgA8Hpo6KptDCWGs9XAZBl6qJvk0h+SNr
8Yp4c16Bygo+RilXBDPsRXgHB3o9kVqT/XYPSnurbRtf6S6OL3FAZtw0TMMIG+Kn
HgaVUEp7yPVnz4ymx+yYM88NpEUlqULI/fHcDPNyAwXk5FUKzD1OEzL6GkhrPuqC
XAx7Mf2r+hW2rWhsSLujascA5zB4KcIShaQbOoGtenA6ldlhuKXdFTYZrng=
-----END CERTIFICATE-----`

const fakeHENKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAwI7lSZljbGSeCfvhfphMioC0dLEJCZxXllQ+ZenJo+6+QkdJ
+wiAyiNZbL1psC0+3+ZzpaJqw9Ov64NLK06RxAXbW9biGuQvyF+o17YzMaYX0mRR
KcffioP1UInvQGUFYnNgR7Brjak6qnGrKU4z5dbbZeMvf7I7q2KyewQ1JHwMm1Tt
I/nw3ZUr3fMKtUsHhWlN/brvBiKkPiG0E1XuVogpHxJY5AyipxKjtHInsqBjAOw2
BJd9YAvFiISoRmT51Q3xxCRSDTEjPLSk3oZlleTVo/aUZn8n34skaXhVBjDS6rYC
daPDRHpsg4cw1JaamdDpYbIIjgs+86aY3zRQuwIDAQABAoIBAFzrV7e5XiHrN9wn
gPv+8EiRrQL2fx71I8r2Iho5w8Toq0T+c7PAua1Re5CeooaSftm6sinGg3C2ERk8
BSUDyBFoph7eRcOmQ2yUxLw7Pt8BgFNVd1kLC1MjNFjGBv7zALMua9KMTopQlG+1
ZFwNUbvif4LeK4iacHLWsLvuHtrYbX6BeE8+lVGrNhmwa+bBDq5vepmI2PYQPl2b
RAOKv9RA/bXc8D2TWGBdOzo3YLl0kI9GZVq+xNRSnpKtKn5qlmkfZ6zAASutDEU/
gJMr164cBuMg/L6q/mDWaGl5wMW1sVsB6CdUzeJs7FBkRvAkirdrSFwiDcj4ctbi
kq44nRkCgYEA/PB1XLO3UBmXgSnRvh1E1cj2BsrhaLGj+lMD4b0iixB34zI7P1jV
Y2TGR4MRMYBmzjY1bWr7QAgbK68YHA5koSShZ/0OYzRZdZs7+nYib5wKV+DuEXnt
cSxTuHIILvPzk4rtSQqgs2rFnFQiCR+vPfVRhe+WuaXK3dQeLFlhxLUCgYEAwuNk
VFxbGi+8v3Wsis89VTx4IGyUIa3kjOAR6sr4lW3upbqhcjoZveGTccUpaDrX9xS5
xgp52RaBZQUU62Kk+Rlqc4xFYKvl5r7l5BxPw+8W9bwKUU4t3PlT6/akuQvAYj/O
qBGc1pf2Hdg9ttpQd3oztAr7nhU3b6SFIgDMFa8CgYEAlyoiCdACCyXwTKowhp05
aUbb+j0/r3ES3eTFGiENxuyFqct4ayhtByTP9ycWnG3vgugU0Bqyo5b0ngvbrdDQ
RRn+OIadFZ7QpB+tHceCVw97gv+TZ/BlflCOjFniGCWFebT6kL+AQRnblc0WNjuw
YKf/G7uPac3yytYdXkXgz00CgYAjg9rZwMbdW8uyvFgIJ8IOkWl2xzKrfIwE3CSH
vBtW5+SwkPUw4sOkJcJ/3iUwmGCY507/dxNDa2WDKkzopF5aArayfeJ6vniz9x/f
1QT4OM7fUzgyHuQeu9T+UEEuc6TIgsY/PI5vUNwKDkkY1GoLi9p2OfYmlck3cCzO
yIRogwKBgQDfS7F9db42UnCUumwLfySvcjUd4vVmnN/cpxI5SrxaPwkudiDXgzpU
PCVId1PsX2rDTRr7vVM+j9nDtEybYZO7BFHPGPRkkHr30mU17VSOxWsfnN3+NJ3u
0lXh31j77WmmE69C/xM0M0ztKHkrQR/UIO0twl8SGIcLHEak/XIp0A==
-----END RSA PRIVATE KEY-----`

func TestCreateServerConfig(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		t.Run("WrongCACrtFmt", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			os.RemoveAll(fakeCertsPath)

			if _, err := createServerConfig(fakeCertsPath); err == nil {
				t.Error(unexpectedSuccess)
			}

			if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte("hello"), 0444); err != nil {
				t.Error(err.Error())
			}

			if _, err := createServerConfig(fakeCertsPath); err == nil {
				t.Error(unexpectedSuccess)
			}
		})
		t.Run("AbsentHenCrtKey", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(fakeCASert), 0444); err != nil {
				t.Error(err.Error())
			}
			if _, err := createServerConfig(fakeCertsPath); err == nil {
				t.Error(unexpectedSuccess)
			}
		})
	})

	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakeCertsPath)

		err := os.MkdirAll(fakeCertsPath, os.ModePerm)
		if err != nil {
			t.Error(err.Error())
		}

		if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(fakeCASert), 0444); err != nil {
			t.Error(err.Error())
		}

		if err := os.WriteFile(fakeCertsPath+"/hen-crt.pem", []byte(fakeHENSert), 0444); err != nil {
			t.Error(err.Error())
		}
		if err := os.WriteFile(fakeCertsPath+"/hen-key.pem", []byte(fakeHENKey), 0444); err != nil {
			t.Error(err.Error())
		}
		if _, err := createServerConfig(fakeCertsPath); err != nil {
			t.Error(unexpectedFail)
		}

	})
}

const addr = "localhost:12345"

func TestListenAndServe(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer os.RemoveAll(fakeCertsPath)

		if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
			t.Error(err.Error())
		}
		if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(fakeCASert), 0444); err != nil {
			t.Error(err.Error())
		}
		if err := os.WriteFile(fakeCertsPath+"/hen-crt.pem", []byte(fakeHENSert), 0444); err != nil {
			t.Error(err.Error())
		}
		if err := os.WriteFile(fakeCertsPath+"/hen-key.pem", []byte(fakeHENKey), 0444); err != nil {
			t.Error(err.Error())
		}

		s := TLSServer{Certspath: fakeCertsPath}
		go func() {
			time.Sleep(2 * time.Second)
			s.listener.Close()
		}()
		s.ListenAndServe(addr, nil)
		time.Sleep(1 * time.Second)

	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("AbsentCertsPath", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			s := TLSServer{Certspath: fakeCertsPath}
			go func() {
				time.Sleep(2 * time.Second)
				s.listener.Close()
			}()
			s.ListenAndServe(addr, nil)
			time.Sleep(1 * time.Second)
		})
		t.Run("WrongCACrtFmt", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(""), 0444); err != nil {
				t.Error(err.Error())
			}

			s := TLSServer{Certspath: fakeCertsPath}
			go func() {
				time.Sleep(2 * time.Second)
				s.listener.Close()
			}()
			s.ListenAndServe(addr, nil)
			time.Sleep(1 * time.Second)
		})
		t.Run("AbsentHenCrtKey", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(fakeCASert), 0444); err != nil {
				t.Error(err.Error())
			}

			s := TLSServer{Certspath: fakeCertsPath}
			go func() {
				time.Sleep(2 * time.Second)
				s.listener.Close()
			}()
			s.ListenAndServe(addr, nil)
			time.Sleep(1 * time.Second)
		})
		t.Run("BadAddr", func(t *testing.T) {
			defer func() {
				os.RemoveAll(fakeCertsPath)
				if r := recover(); r == nil {
					t.Error(r)
				}
			}()

			if err := os.MkdirAll(fakeCertsPath, os.ModePerm); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/ca-crt.pem", []byte(fakeCASert), 0444); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/hen-crt.pem", []byte(fakeHENSert), 0444); err != nil {
				t.Error(err.Error())
			}
			if err := os.WriteFile(fakeCertsPath+"/hen-key.pem", []byte(fakeHENKey), 0444); err != nil {
				t.Error(err.Error())
			}

			s := TLSServer{Certspath: fakeCertsPath}
			go func() {
				time.Sleep(2 * time.Second)
				s.listener.Close()
			}()
			s.ListenAndServe("_", nil)
			time.Sleep(1 * time.Second)
		})
	})
}
