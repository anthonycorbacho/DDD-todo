package database

import "testing"

func TestOpen(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		name       string
		driver     string
		connection string
		withError  bool
	}{
		{
			name:       "unknown driver should fail",
			driver:     "lolodb",
			connection: "aa@bb://dddd/ww",
			withError:  true,
		},
		{
			name:       "mysql driver should pass",
			driver:     "mysql",
			connection: "username:password@tcp(host.mysql.tld:3306)/dbname",
			withError:  false,
		},
		{
			name:       "mysql driver should fail with invalid connection",
			driver:     "mysql",
			connection: "host.mysql.tld:3306/dbname",
			withError:  true,
		},
		{
			name:       "postgres driver should pass",
			driver:     "postgres",
			connection: "user:password@host:4242/dbname?useSSl=true",
			withError:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := Open(tc.driver, tc.connection)
			if tc.withError != (err != nil) {
				t.Errorf("[%s] fail : %v", tc.name, err)
			}
		})
	}
}
