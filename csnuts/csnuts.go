// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package csnuts

import (
//	"io"
	"net/http"
	"text/template"
//	"strconv"
	"appengine"
	"appengine/datastore"
	"appengine/user"

)

const lenSummery=1370
//var Site string="http://www.csnuts.com"
var Site string="http://localhost:8080"

type TagCount struct {
        TagName string
        Count int
}
type listPage struct {
    U *user.User
    Loginbar string
    Tag string
    SiteBase string
    QueryBase string
    CurrPageBase string
    Msgs []*Message
    Arts []*Article
    NumMsgs Count
    TagCloud []*TagCount
}
type articlePage struct {
    U *user.User
    IsAdmin bool
    Loginbar string
    SiteBase string
    CurrPageBase string
    Msg *Message
    Art *Article
    Cmts []*Comment
    TagCloud []*TagCount
}
func handleMainPage(w http.ResponseWriter, r *http.Request) {
	var pageData listPage
    pageData.SiteBase=Site
    pageData.QueryBase=Site+"/query/?"
	pageData.NumMsgs.Value=getCount(w,r)
	if r.Method != "GET" || r.URL.Path != "/" {
		serve404(w)
		return
	}
	c := appengine.NewContext(r)
	u:=user.Current(c)
    pageData.U=u
	if u==nil {
		url,_:=user.LoginURL(c,"/")
		pageData.Loginbar="<a href=\""+url+"\">登入</a>"
	} else {
		url,_:=user.LogoutURL(c,"/")
		pageData.Loginbar="欢迎,"+u.String()+"(<a href=\""+url+"\">登出</a>)"
	}
	q := datastore.NewQuery("Message").Order("-Date").Limit(10)
	ks, err := q.GetAll(c, &(pageData.Msgs))
	if err != nil {
		serveError(c, w, err)
		return
	}
    // display ID from Key.IntID
    for i,_:=range pageData.Msgs {
        pageData.Msgs[i].ID=ks[i].IntID()
        pageData.Msgs[i].Content=[]byte(SubstrByByte(string(pageData.Msgs[i].Content),lenSummery))
    }
    // tagcloud
    tags:=new([]*Tag)
	q = datastore.NewQuery("Tag").Order("-Count").Limit(100)
	ks, err = q.GetAll(c, tags)
	if err != nil {
		serveError(c, w, err)
		return
	}
    for i,k:=range ks {
        tagcount:=new(TagCount)
        tagcount.TagName=k.StringID()
        tagcount.Count=(*tags)[i].Count
        pageData.TagCloud=append(pageData.TagCloud, tagcount)
    }
    //end tagcloud
    pageData.Arts=Msgs2Arts(pageData.Msgs)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	mainPage,err:= template.ParseFiles("template/msglist.html","template/articles.html","template/header.html","template/footer.html")
	if(err!=nil) {
		c.Errorf("%v", err)
		return
	}
	if err := mainPage.Execute(w, pageData); err != nil {
		c.Errorf("%v", err)
	}
}

func init() {
    //http.HandleFunc("/tag/good/",handleMsgGood)
//    http.HandleFunc("/msg/good/",handleMsgGood)
	http.HandleFunc("/", handleMainPage)
	http.HandleFunc("/sign", handleSign)
    http.HandleFunc("/comment",handleComment)
    http.HandleFunc("/tagquery/",handleTagQuery)
    http.HandleFunc("/tag/",handleTaggedMsgs)

    http.HandleFunc("/msg/" ,handleMsg      )
	http.HandleFunc("/query/",handleMsgQuery )
    http.HandleFunc("/del/" ,handleMsgDelete)
    http.HandleFunc("/good/",handleMsgGood  )
}
