package store

import (
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoStore struct {
	Conn string
	Name string
}

func (self *MongoStore) Init() error {
	if len(self.Conn) == 0 || len(self.Name) == 0 {
		return errors.New("Mongo store configuration error")
	}

	session, errConnect := mgo.Dial(self.Conn)
	if errConnect != nil {
		return errConnect
	}
	defer session.Close()

	return nil
}

func (self *MongoStore) DeleteObjects(dataset string, oids []string) (count int64, err error) {
	if len(oids) == 0 {
		return 0, nil
	}

	queries := []bson.M{}
	for _, oid := range oids {
		if bson.IsObjectIdHex(oid) {
			query := bson.M{"_id": bson.ObjectIdHex(oid)}
			queries = append(queries, query)
		}
	}

	filter := bson.M{"$or": queries}

	session, errConnect := mgo.Dial(self.Conn)
	if errConnect != nil {
		return 0, errConnect
	}
	defer session.Close()

	c := session.DB(self.Name).C(dataset)

	info, errDelete := c.RemoveAll(filter)
	if errDelete != nil {
		return 0, errDelete
	}

	return int64(info.Removed), nil
}

func (self *MongoStore) Destroy() error {
	return nil
}
