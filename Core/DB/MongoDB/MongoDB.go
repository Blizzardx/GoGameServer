package MongoDB

import (
	"github.com/davyxu/golog"
	"gopkg.in/mgo.v2"
	"litgame.cn/Server/Core/Common"
)

type MongoDBConnector struct {
	dbAdd     string
	db        string
	log       *golog.Logger
	dbSession *mgo.Session
}

func New(add string, dbName string) *MongoDBConnector {
	conn := &MongoDBConnector{dbAdd: add, db: dbName, log: golog.New("db.MongoDB")}
	if !conn.ConnectToDb() {
		return nil
	}
	return conn
}
func (dbConnector *MongoDBConnector) ConnectToDb() bool {
	session, err := mgo.Dial(dbConnector.dbAdd)
	if err != nil {
		dbConnector.log.Errorln("error on connect mongo db ", dbConnector.dbAdd)
		return false
	}
	dbConnector.dbSession = session
	dbConnector.dbSession.SetMode(mgo.Monotonic, true)
	return true
}
func (dbConnector *MongoDBConnector) DoSth(getSession func(session *mgo.Session)) {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	Common.SafeCall(func() {
		getSession(s)
	})
}
func (dbConnector *MongoDBConnector) Insert(collation string, value interface{}) error {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	err := s.DB(dbConnector.db).C(collation).Insert(value)
	if nil != err {
		dbConnector.log.Errorln("error on Insert to mongo db ", err, dbConnector.db, collation, value)
		return err
	}
	return nil
}
func (dbConnector *MongoDBConnector) Delete(collation string, query interface{}) error {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	collection := s.DB(dbConnector.db).C(collation)
	err := collection.Remove(query)
	if nil != err {
		dbConnector.log.Errorln("error on Delete from mongo db ", err, dbConnector.db, collation, query)
		return err
	}
	return nil
}
func (dbConnector *MongoDBConnector) Update(collation string, query interface{}, value interface{}) error {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	collection := s.DB(dbConnector.db).C(collation)
	err := collection.Update(query, value)
	if nil != err {
		dbConnector.log.Errorln("error on Update from mongo db ", err, dbConnector.db, collation, query, value)
		return err
	}
	return nil
}
func (dbConnector *MongoDBConnector) Upsert(collation string, query interface{}, value interface{}) error {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	collection := s.DB(dbConnector.db).C(collation)
	_, err := collection.Upsert(query, value)
	if nil != err {
		dbConnector.log.Errorln("error on Upsert from mongo db ", err, dbConnector.db, collation, query, value)
		return err
	}
	return nil
}
func (dbConnector *MongoDBConnector) Find(collation string, query interface{}, queryObjectInstance interface{}) (bool, error) {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	con := s.DB(dbConnector.db).C(collation)
	if err := con.Find(query).One(queryObjectInstance); err != nil {
		if err.Error() != mgo.ErrNotFound.Error() {
			dbConnector.log.Errorln("error on Find from mongo db ", err, dbConnector.db, collation, query)
			return false, err
		} else {
			return false, nil
		}
	}
	return true, nil
}
func (dbConnector *MongoDBConnector) EnsureIndex(collation string, index mgo.Index) error {
	s := dbConnector.dbSession.Copy()
	defer s.Close()
	con := s.DB(dbConnector.db).C(collation)

	err := con.EnsureIndex(index)
	if nil != err {
		dbConnector.log.Errorln("error on EnsureIndex from mongo db ", err, dbConnector.db, collation, index)
	}
	return err
}
