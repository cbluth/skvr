package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func api(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	k, ns := apiArgs(r)
	switch {
	case r.Method == http.MethodGet:
		{
			apiGet(k, ns, w, r)
		}
	case r.Method == http.MethodOptions:
		{
			apiOpt(k, ns, w, r)
		}
	case (r.Method == http.MethodPost || r.Method == http.MethodPut):
		{
			apiPut(k, ns, w, r)
		}
	case r.Method == http.MethodDelete:
		{
			apiDel(k, ns, w, r)
		}
	default:
		{
			// log.Println("default case hit")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func apiArgs(r *http.Request) (key, namespace) {
	k := key{}
	ns := namespace{}
	path := strings.Split(r.URL.Path, "/")[1:]
	switch {
	// "/"
	case len(path) == 1 && path[0] == "":
		{
			// do nothing: return empty key and ns
		}
	// "/<key>"
	case len(path) == 1 && path[0] != "":
		{
			ns = defaultNamespace
			k = key(path[0])
		}
	// "/<namespace>/"
	case len(path) == 2 && path[0] != "" && path[1] == "":
		{
			ns = namespace(path[0])
		}
	// "/<namespace>/<key>"
	case len(path) >= 2 && path[0] != "" && path[1] != "":
		{
			ns = namespace(path[0])
			k = key(path[1])
		}
	}
	// log.Println(k, ns)
	return k, ns
}

func apiGet(k key, ns namespace, w http.ResponseWriter, r *http.Request) {
	nn := namespace{}
	nk := key{}
	switch {
	case k.String() == "" && ns.String() == "":
		{
			nn = defaultNamespace
			nk = defaultKey
		}
	case k.String() == "" && ns.String() != "":
		{
			nn = ns
			nk = defaultKey
		}
	case k.String() != "" && ns.String() != "":
		{
			nn = ns
			nk = k
		}
	}
	v, err := nn.kvGet(nk)
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			apiOpt(k, ns, w, r)
			return
		}
		log.Println("ERROR "+r.Method+":", err)
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%s: %s :: %s", r.Method, nn.String(), nk.String())
	w.WriteHeader(http.StatusOK)
	w.Write(v)
}

func apiOpt(k key, ns namespace, w http.ResponseWriter, r *http.Request) {
	switch {
	case k.String() == "" && ns.String() == "":
		{
			nss := getBuckets()
			log.Println("List Namespaces:", nss)
			w.Header().Set("List", "Namespaces")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("List Namespaces:\n"))
			w.Write([]byte("---\n"))
			for _, ns := range nss {
				w.Write([]byte(ns + "\n"))
			}
		}
	case k.String() == "" && ns.String() != "":
		{
			keys, err := ns.listKeys()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("List", "Keys")
			log.Println("List: "+ns.String()+" ::", keys)
			if len(keys) == 0 {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("List Keys:\n"))
    			w.Write([]byte("---\n"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("List Keys:\n"))
    			w.Write([]byte("---\n"))
				for _, ki := range keys {
					w.Write([]byte(ki + "\n"))
				}
			}
		}
	case k.String() != "":
		{
			_, err := ns.kvGet(k)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
}

func apiPut(k key, ns namespace, w http.ResponseWriter, r *http.Request) {
	switch {
	case k.String() == "" && ns.String() == "":
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	case k.String() == "" && ns.String() != "":
		{
			err := ns.kvPut(nil, nil)
			if err != nil {
				log.Println("ERROR "+r.Method+":", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	case k.String() != "" && ns.String() != "":
		{
			v, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = ns.kvPut(k, v)
			if err != nil {
				log.Println("ERROR "+r.Method+":", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s :: %s", r.Method, ns.String(), k.String())
			w.WriteHeader(http.StatusOK)
		}
	}
}

func apiDel(k key, ns namespace, w http.ResponseWriter, r *http.Request) {
	switch {
	case k.String() == "" && ns.String() == "":
		{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	case k.String() == "" && ns.String() != "":
		{
			err := ns.delete()
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Printf("%s Namespace: %s", r.Method, ns.String())
			w.WriteHeader(http.StatusOK)
		}
	case k.String() != "" && ns.String() != "":
		{
			err := ns.kvDelete(k)
			if err != nil {
				log.Println("ERROR "+r.Method+":", err)
				if strings.Contains(err.Error(), "not found") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s :: %s", r.Method, ns.String(), k.String())
			w.WriteHeader(http.StatusOK)
		}
	}
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || len(strings.TrimSpace(u)) < 1 || len(strings.TrimSpace(p)) < 1 {
				unauthorised(w)
				return
			}
			// This is a dummy check for credentials.
			if u != "hello" || p != "world" {
				unauthorised(w)
				return
			}
			// If required, Context could be updated to include authentication
			// related data so that it could be used in consequent steps.
			next(w, r)
		},
	)
}

func unauthorised(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	w.WriteHeader(http.StatusUnauthorized)
}

