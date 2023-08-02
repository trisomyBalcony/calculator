package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Expression struct {
	Expression string `json:"expression"`
}

type Solution struct {
	Solution int `json:"solution"`
}

type Error struct {
	Error string `json:"error"`
}

func checkAccess(r *http.Request) bool {
	return r.Header.Get("User-Access") == "superuser"
}

func parseExpression(exp string) (int, error) {
	exp = strings.ReplaceAll(exp, " ", "")
	fmt.Printf("exp: %v\n", exp)
	if strings.ContainsAny(exp, "*/") {
		return 0, fmt.Errorf("неподдерживаемая операция в выражении: %s", exp)
	}

	re := regexp.MustCompile(`(\+|-)?\d+`)
	tokens := re.FindAllString(exp, -1)
	fmt.Printf("tokens: %v\n", tokens)
	result := 0
	for _, token := range tokens {
		if num, err := strconv.Atoi(token); err == nil {
			result += num
		} else {
			return 0, fmt.Errorf("некорректное число: %s", token)
		}
	}

	return result, nil
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if !checkAccess(r) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Println("Доступ запрещён")
		json.NewEncoder(w).Encode(Error{Error: "Доступ запрещён"})
		return
	}

	var expression Expression

	if r.Method == "POST" {
		err := json.NewDecoder(r.Body).Decode(&expression)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("Ошибка при чтении JSON", err)
			return
		}
	} else if r.Method == "GET" {
		expression.Expression = r.URL.Query().Get("expression")
		expression.Expression = strings.ReplaceAll(expression.Expression, " ", "+")
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Println("Метод не поддерживается")
		return
	}

	solution, err := parseExpression(expression.Expression)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("Ошибка при вычислении выражения", err)
		json.NewEncoder(w).Encode(Error{Error: fmt.Sprintf("Ошибка при вычислении выражения: %s", err.Error())})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Solution{Solution: solution})
}

func main() {
	http.HandleFunc("/calculate", calculateHandler)
	http.ListenAndServe(":8080", nil)
}
