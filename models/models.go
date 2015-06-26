package models

import (
	"github.com/Unknwon/com"
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	_DB_NAME        = "data/beeblog.db"
	_SQLITE3_DRIVER = "sqlite3"
)

type Category struct {
	Id              int64
	Title           string
	Created         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	TopicTime       time.Time `orm:"index"`
	TopicCount      int64
	TopicLastUserId int64
}

type Topic struct {
	Id              int64
	Uid             int64
	Title           string
	Category        string
	Labels          string
	Content         string `orm:"size(5000)"`
	Attachment      string
	Created         time.Time `orm:"index"`
	Updated         time.Time `orm:"index"`
	Views           int64     `orm:"index"`
	Author          string
	ReplyTime       time.Time `orm:"index"`
	ReplyCount      int64
	ReplyLastUserId int64
}

type Comment struct {
	Id      int64
	Tid     int64
	Name    string
	Content string    `orm:"size(1000)"`
	Created time.Time `orm:"index"`
}

func RegisterDB() {
	if !com.IsExist(_DB_NAME) {
		os.MkdirAll(path.Dir(_DB_NAME), os.ModePerm)
		os.Create(_DB_NAME)
	}

	orm.RegisterModel(new(Category), new(Topic), new(Comment))
	orm.RegisterDriver(_SQLITE3_DRIVER, orm.DR_Sqlite)
	orm.RegisterDataBase("default", _SQLITE3_DRIVER, _DB_NAME, 10)
}

func AddReply(tid, nickname, content string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}

	reply := &Comment{
		Tid:     tidNum,
		Name:    nickname,
		Content: content,
		Created: time.Now(),
	}

	o := orm.NewOrm()
	_, err = o.Insert(reply)

	if err != nil {
		return err
	}

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		topic.ReplyTime = time.Now()
		topic.ReplyCount++
		_, err = o.Update(topic)
	}

	return err
}

func AddTopic(title, category, label, content, attachment string) error {
	//处理标签
	label = "$" + strings.Join(
		strings.Split(label, " "), "#$") + "#"

	//	空格作为多个标签的分隔符
	//	beego
	//	$beego#
	//	orm
	//	$beego#$bee#
	//	$bee#

	o := orm.NewOrm()

	topic := &Topic{
		Title:      title,
		Category:   category,
		Labels:     label,
		Content:    content,
		Attachment: attachment,
		Created:    time.Now(),
		Updated:    time.Now(),
		ReplyTime:  time.Now(),
	}

	_, err := o.Insert(topic)

	if err != nil {
		return err
	}

	//更新分类统计
	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("title", category).One(cate)
	if err == nil {
		//如果不存在 简单的忽略更新操作
		cate.TopicCount++
		_, err = o.Update(cate)
	}

	return err
}

func AddCategory(name string) error {
	o := orm.NewOrm()
	cate := &Category{Title: name}

	qs := o.QueryTable("category")
	err := qs.Filter("title", name).One(cate)
	if err == nil {
		return err
	}

	curTime := time.Now()

	cate.Created = curTime
	cate.TopicTime = curTime

	_, err = o.Insert(cate)
	if err != nil {
		return err
	}
	return nil
}

func GetAllTopics(cate string, label string, isDesc bool) (topics []*Topic, err error) {
	o := orm.NewOrm()
	topics = make([]*Topic, 0)
	qs := o.QueryTable("topic")

	if isDesc {
		if len(cate) > 0 {
			qs = qs.Filter("category", cate)
		}

		if len(label) > 0 {
			qs = qs.Filter("labels__contains", "$"+label+"#")
		}

		_, err = qs.OrderBy("-created").All(&topics)
	} else {
		_, err = qs.All(&topics)

	}

	return topics, err

}

func GetAllCategories() ([]*Category, error) {
	o := orm.NewOrm()
	cates := make([]*Category, 0)
	qs := o.QueryTable("category")

	_, err := qs.All(&cates)
	return cates, err
}

func GetTopic(tid string) (*Topic, error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()
	topic := new(Topic)
	qs := o.QueryTable("topic")
	err = qs.Filter("id", tidNum).One(topic)
	if err != nil {
		return nil, err
	}

	topic.Views++
	_, err = o.Update(topic)

	topic.Labels = strings.Replace(strings.Replace(topic.Labels, "#", " ", -1), "$", "", -1)

	return topic, err
}

func ModifyTopic(tid, title, category, label, content, attachment string) error {

	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}

	//处理标签
	label = "$" + strings.Join(
		strings.Split(label, " "), "#$") + "#"

	var oldCate, oldAttach string

	o := orm.NewOrm()
	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		oldCate = topic.Category
		oldAttach = topic.Attachment
		topic.Title = title
		topic.Category = category
		topic.Labels = label
		topic.Content = content
		topic.Attachment = attachment
		topic.Updated = time.Now()
		_, err = o.Update(topic)
		if err != nil {
			return err
		}
	}

	//更新分类统计
	if len(oldCate) > 0 {
		cate := new(Category)
		qs := o.QueryTable("category")
		err := qs.Filter("title", oldCate).One(cate)
		if err == nil {
			println(".........old cate category: ", cate)

			cate.TopicCount--
			_, err = o.Update(cate)
		}
	}

	// 删除旧的附件
	if len(oldAttach) > 0 {
		os.Remove(path.Join("attachment", oldAttach))
	}

	cate := new(Category)
	qs := o.QueryTable("category")
	err = qs.Filter("title", category).One(cate)
	if err == nil {
		println(".........new cate category: ", cate)

		cate.TopicCount++
		_, err = o.Update(cate)
	}

	return nil
}

func DelCategory(id string) error {
	cid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	cate := &Category{Id: cid}
	_, err = o.Delete(cate)
	return err

}

func DeleteTopic(tid string) error {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return err
	}
	var oldCate string
	o := orm.NewOrm()
	topic := &Topic{Id: tidNum}

	if o.Read(topic) == nil {
		oldCate = topic.Category
		_, err = o.Delete(topic)
		if err != nil {
			return err
		}
	}

	if len(oldCate) > 0 {
		cate := new(Category)
		qs := o.QueryTable("category")
		err = qs.Filter("title", oldCate).One(cate)
		if err == nil {
			cate.TopicCount--
			_, err = o.Update(cate)
		}
	}

	_, err = o.Delete(topic)
	return err
}

func GetAllReplies(tid string) (replies []*Comment, err error) {
	tidNum, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	replies = make([]*Comment, 0)

	o := orm.NewOrm()
	qs := o.QueryTable("comment")
	_, err = qs.Filter("tid", tidNum).All(&replies)
	return replies, err
}

func DeleteReply(rid string) error {
	ridNum, err := strconv.ParseInt(rid, 10, 64)
	if err != nil {
		return err
	}

	o := orm.NewOrm()

	var tidNum int64

	reply := &Comment{Id: ridNum}

	if o.Read(reply) == nil {
		tidNum = reply.Tid
		_, err = o.Delete(reply)
		if err != nil {
			return err
		}
	}

	replices := make([]*Comment, 0)

	qs := o.QueryTable("comment")
	_, err = qs.Filter("tid", tidNum).OrderBy("-created").All(&replices)
	if err != nil {
		return err
	}

	topic := &Topic{Id: tidNum}
	if o.Read(topic) == nil {
		if len(replices) > 0 {
			topic.ReplyTime = replices[0].Created
		} else {
			topic.ReplyTime = topic.Created
		}
		topic.ReplyCount = int64(len(replices))
		_, err = o.Update(topic)
	}

	/*
		1. 1 2 3
	*/

	return err
}
