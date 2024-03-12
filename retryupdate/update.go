//go:build !solution

package retryupdate

import (
	"errors"
	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/retryupdate/kvapi"
)

func UpdateValue(c kvapi.Client, key string, updateFn func(oldValue *string) (newValue string, err error)) error {
	newVal := ""
Restart_:
	getResp, err := c.Get(&kvapi.GetRequest{Key: key})
OuterLoop:
	for {
		if errors.Is(err, kvapi.ErrKeyNotFound) {
			newVal, err = updateFn(nil)
			if err != nil {
				return err
			}
		Loop:
			for {
				_, err = c.Set(&kvapi.SetRequest{Key: key, Value: newVal, NewVersion: uuid.Must(uuid.NewV4())})
				auth := kvapi.AuthError{}
				authp := &auth
				conf := kvapi.ConflictError{}
				confp := &conf
				switch {
				case errors.As(err, &authp):
					return err
				case err == nil:
					break Loop
				case errors.As(err, &confp):
					continue OuterLoop
				}
			}
			return nil
		}
		auth := kvapi.AuthError{}
		authp := &auth
		switch {
		case errors.As(err, &authp):
			return err
		case err != nil:
			goto Restart_
		}
		newVal, err = updateFn(&getResp.Value)
		if err != nil {
			return err
		}
		newV := uuid.Must(uuid.NewV4())
	Loop2:
		for {
			_, err = c.Set(&kvapi.SetRequest{Key: key, Value: newVal, OldVersion: getResp.Version, NewVersion: newV})
			auth := kvapi.AuthError{}
			authp := &auth
			conf := kvapi.ConflictError{}
			confp := &conf
			switch {
			case errors.As(err, &authp):
				return err
			case err == nil:
				break Loop2
			case errors.As(err, &confp):
				if err.(*kvapi.APIError).Unwrap().(*kvapi.ConflictError).ExpectedVersion == newV {
					return nil
				}
				goto Restart_
			case errors.Is(err, kvapi.ErrKeyNotFound):
				continue OuterLoop
			}
		}
		return nil
	}
}
