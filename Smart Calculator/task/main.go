package main

/*
[Smart Calculator - Stage 7/7: I've got the power](https://hyperskill.org/projects/74/stages/415/implement)
-------------------------------------------------------------------------------
[Stack](https://hyperskill.org/learn/step/5252)
[Math package](https://hyperskill.org/learn/topic/2012)
[Using stack in Go] -- TODO!
*/

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type ExpressionType int

const (
	_ ExpressionType = iota
	Number
	Symbol
	Variable
)

type OperationType int

const (
	_ OperationType = iota
	Assignment
	Regular
)

type Expression struct {
	ExpressionType
	Value any
}

type Calculator struct {
	memory      map[string]int
	stack       []Expression
	postfixExpr []Expression
	expression  []Expression
	OperationType
}

var opRank = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
	"^": 3,
}

var symbols = []string{"+", "-", "=", "*", "/", "(", ")", "^"}

// mapContains checks if a map contains a specific element
func mapContains(m map[string]int, key string) bool {
	_, ok := m[key]
	return ok
}

// sliceContains checks if a slice contains a specific element
func sliceContains(s []string, element string) bool {
	for _, x := range s {
		if x == element {
			return true
		}
	}
	return false
}

// isNumeric checks if all the characters in the string are numbers
func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// isAlpha checks if all the characters in the string are alphabet letters
func isAlpha(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

func isSymbol(token string) bool {
	return sliceContains(symbols, token)
}

// pop deletes the last element of the stack []Expression and returns it
func pop(alist *[]Expression) Expression {
	f := len(*alist)
	rv := (*alist)[f-1]
	*alist = (*alist)[:f-1]
	return rv
}

// checkAssignment checks if the line is an assignment operation "a = 5"
func checkAssignment(s string) bool {
	if strings.Contains(s, "=") && strings.Count(s, "=") == 1 {
		return true
	}
	return false
}

// getAssignmentElements returns the elements of an assignment operation "a = 5"
func getAssignmentElements(line string) []Expression {
	var elems []Expression
	var end int
	var number any
	var variable string

	for len(line) > 0 {
		token := string(line[0])
		switch token {
		case " ":
			end = 1
		case "=":
			end = 1
			elems = append(elems, Expression{Symbol, token})
		default:
			if isNumeric(token) {
				number, end = parseNumber(line)
				elems = append(elems, Expression{Number, number})
			}
			if isAlpha(token) {
				variable, end = parseVariable(line)
				elems = append(elems, Expression{Variable, variable})
			}
		}
		line = line[end:]
	}
	return elems
}

// The assign function assigns a value to a variable and stores it in the calculator memory
func (c Calculator) assign(line string) {
	elems := getAssignmentElements(line)
	if elems == nil {
		return
	}

	variable := elems[0].Value
	value := elems[2].Value

	if fmt.Sprintf("%T", value) == "string" {
		value = c.getVarValue(value.(string))
		if value == nil {
			return
		}
	}
	c.memory[variable.(string)] = value.(int)
	return
}

// processCommand checks if the input is command is either "/exit" or "/help" and if not reports an error.
func processCommand(line string) {
	if line != "/exit" && line != "/help" {
		fmt.Println("Unknown command")
		return
	}
}

// checkParenthesis checks if there are the same amount of parenthesis on both sides of the infixExpr
func checkParenthesis(line string) bool {
	return strings.Count(line, "(") != strings.Count(line, ")")
}

// checkSymbols checks if the expression has any valid symbols and that it isn't
// an invalid expression like 10 10 or 10 10 * 10
func checkSymbols(line string) bool {
	for _, symbol := range symbols {
		if strings.Count(line, symbol) > 0 {
			return true
		}
	}
	return false
}

func getOperationType(line string) OperationType {
	if checkAssignment(line) {
		return Assignment
	}
	return Regular
}

func validateExpression(line string) bool {
	var number, end int
	var varName string

	// First check if the expression is a single number or a single variable
	if isNumeric(line) || isAlpha(line) {
		return true
	}

	// Then check for the most basic case of invalid expressions, trailing operators like: 10+10-8-
	if strings.HasSuffix(line, "+") || strings.HasSuffix(line, "-") {
		fmt.Println("Invalid expression")
		return false
	}

	// Then check if the line has more than one "=" sign in it
	if strings.Count(line, "=") > 1 {
		fmt.Println("Invalid assignment")
		return false
	}

	// Then check if there are the same amount of parenthesis on both sides of the line
	if checkParenthesis(line) {
		fmt.Println("Invalid expression")
		return false
	}

	// Then check if there is at least one valid symbol in the line, to validate cases like 10 10 or 18 22
	// For cases like a2a or n22 that begin with a letter, then we should print "Invalid identifier" instead
	// So for cases that start with a letter, like a2a we return true and further check within validateSyntax()
	if !checkSymbols(line) && !isAlpha(line[0:1]) {
		fmt.Println("Invalid expression")
		return false
	}

	// If none of the above checks are true, then we perform the final check,
	// We proceed to validate the syntax of the expression:
	return validateSyntax(line, end, number, varName)
}

// validateSyntax validates the syntax of the expression and checks for special edge cases
func validateSyntax(line string, end int, number any, varName string) bool {
	var prevSym string

	// validateSyntax checks if the expression has any "Invalid identifiers" like a2a or a1 = 8
	// And other edge cases like test = 2n or test = n2
	for len(line) > 0 {
		token := string(line[0])
		switch {
		case token == " ":
			end = 1
		case token == "=":
			end = 1
			prevSym = "="
		case token == "*":
			if prevSym == "*" {
				fmt.Println("Invalid expression")
				return false
			}
			end = 1
			prevSym = token
		case token == "/":
			if prevSym == "/" {
				fmt.Println("Invalid expression")
				return false
			}
			end = 1
			prevSym = token
		case isNumeric(token):
			number, end = parseNumber(line)
			if number == nil && prevSym == "=" { // Validates cases like test = 2n
				fmt.Println("Invalid assignment")
				return false
			}

			if varName == "" && number != nil && prevSym == "=" { // Validates cases like 5 = 5, or 100 = 20
				fmt.Println("Invalid assignment")
				return false
			}
		case isAlpha(token):
			varName, end = parseVariable(line)
			if varName == "" && prevSym == "=" { // Validates cases like test = a2a
				fmt.Println("Invalid assignment")
				return false
			}

			if varName == "" { // Validates cases like a2a or n22 or a1 = 8
				fmt.Println("Invalid identifier")
				return false
			}
		case isSymbol(token):
			_, end = parseSymbol(line)
			prevSym = token
		}
		line = line[end:]
	}
	return true
}

// evalSymbol evaluates the symbol and performs the operation accordingly
func evalSymbol(a, b int, operator any) int {
	switch operator {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		return a / b
	case "^":
		return int(math.Pow(float64(a), float64(b)))
	default:
		return 0
	}
}

// parseNumber parses a number with multiple digits from the input line
func parseNumber(line string) (any, int) {
	var stringNum string
	var end, number int

	for _, t := range line {
		token := string(t)
		if isAlpha(token) {
			return nil, 0
		}

		if !isNumeric(token) {
			break
		}
		stringNum += token
	}
	end = len(stringNum)

	// Convert the string number to an integer number
	number, err := strconv.Atoi(stringNum)
	if err != nil {
		log.Fatal(err)
	}
	return number, end
}

// parseSymbol parses the symbols: "+", "-", "*", "/", "(", ")", "^" from the input line
func parseSymbol(line string) (string, int) {
	var symbol string
	var end int

	for i, t := range line {
		token := string(t)
		if isSymbol(token) {
			symbol += token
			end = i + 1
			break
		}
	}
	end = len(symbol)
	return symbol, end
}

// parseVariable parses a more-than-one-character variable from the input line
func parseVariable(line string) (string, int) {
	var variable string
	var end int

	for _, t := range line {
		token := string(t)
		if isNumeric(token) {
			return "", 0
		}

		if !isAlpha(token) {
			break
		}
		variable += token
	}
	end = len(variable)
	return variable, end
}

// getVarValue returns the value of the variable if it's in the memory of the Calculator
func (c Calculator) getVarValue(variable string) any {
	if !mapContains(c.memory, variable) {
		fmt.Println("Unknown variable")
		return nil
	}
	return c.memory[variable]
}

// appendValues appends the tokens of the line to the c.expression slice
func (c Calculator) appendValues(line string) []Expression {
	var (
		symbol, varName  string
		end              int
		number, varValue any
	)

	for len(line) > 0 {
		token := string(line[0])
		switch {
		case token == " ":
			end = 1
		case isNumeric(token):
			number, end = parseNumber(line)
			c.expression = append(c.expression, Expression{Number, number})
		case isSymbol(token):
			symbol, end = parseSymbol(line)
			c.expression = append(c.expression, Expression{Symbol, symbol})
		case isAlpha(token):
			varName, end = parseVariable(line)
			varValue = c.getVarValue(varName)
			if varValue == nil {
				return nil
			}
			c.expression = append(c.expression, Expression{Number, varValue})
		default:
			return nil
		}
		line = line[end:]
	}
	return c.expression
}

// stackOperator performs the operation on the stack
func (c Calculator) stackOperator(token string) ([]Expression, []Expression) {
	if len(c.stack) == 0 || token == "(" {
		c.stack = append(c.stack, Expression{Symbol, token})
		return c.stack, c.postfixExpr
	}

	if token == ")" {
		for c.stack[len(c.stack)-1].Value != "(" {
			c.postfixExpr = append(c.postfixExpr, pop(&c.stack))
		}
		pop(&c.stack)
		return c.stack, c.postfixExpr
	}

	if higherPrecedence(c.stack[len(c.stack)-1].Value, token) {
		c.stack = append(c.stack, Expression{Symbol, token})
	} else {
		for len(c.stack) > 0 && !higherPrecedence(c.stack[len(c.stack)-1].Value, token) {
			c.postfixExpr = append(c.postfixExpr, pop(&c.stack))
		}
		c.stack = append(c.stack, Expression{Symbol, token})
	}
	return c.stack, c.postfixExpr
}

// getPostfix converts expression to a postfixExpr
func (c Calculator) getPostfix(expression []Expression) []Expression {
	var prevSym any

	for i, token := range expression {
		switch token.ExpressionType {
		case Number:
			if i == 1 && (c.stack[0].Value == "-" || c.stack[0].Value == "+") {
				c.postfixExpr = append(c.postfixExpr, pop(&c.stack))
			}
			c.postfixExpr = append(c.postfixExpr, Expression{Number, token.Value})
			prevSym = token.Value
		case Symbol:
			if sliceContains(symbols, token.Value.(string)) {
				if prevSym == "" || prevSym == nil {
					prevSym = token.Value
					c.stack = append(c.stack, Expression{Symbol, token.Value})
					continue
				}

				if prevSym != token.Value {
					prevSym = token.Value
				} else {
					switch token.Value {
					case "+":
						c.postfixExpr = append(c.postfixExpr, pop(&c.stack))
					case "-":
						pop(&c.stack)
						c.stack = append(c.stack, Expression{Symbol, "-"})
					}
				}
				c.stack, c.postfixExpr = c.stackOperator(token.Value.(string))
			}
		}
	}

	// Append to postfixExpr until c.stack is empty
	for len(c.stack) > 0 {
		c.postfixExpr = append(c.postfixExpr, pop(&c.stack))
	}
	return c.postfixExpr
}

// higherPrecedence returns true if the first symbol has higher precedence than the second
func higherPrecedence(stackPop, token any) bool {
	if stackPop == "(" || opRank[token.(string)] > opRank[stackPop.(string)] {
		return true
	}
	return false
}

// checkSingleNum() checks if the expression is a single number or variable like: --10 or -a or 100
func (c Calculator) checkSingleNum() int {
	var number, numCount, minusCount int

	for i, token := range c.postfixExpr {
		switch token.ExpressionType {
		case Symbol:
			if token.Value.(string) == "-" {
				minusCount += 1
			}
		case Number:
			if numCount > 0 {
				return 0
			}
			number = token.Value.(int)
			numCount += 1
		}

		if i == len(c.postfixExpr)-1 {
			if minusCount%2 != 1 {
				return number
			}
			return number * -1
		}
	}
	return 0
}

// getTotal calculates the total result of postfixExpr
func (c Calculator) getTotal() int {
	var end, minusCount, singleNum int
	var tempMinusCount int

	// If the expression is a single number, return it and then ask the user for the next expression
	singleNum = c.checkSingleNum()
	if singleNum != 0 {
		return singleNum
	}

	// If the expression is not a single number, then we can calculate the total of the postfixExpr
	// Get the first sign "+" or "-" of the original expression
	for _, token := range c.expression {
		if token.ExpressionType != Symbol {
			break
		}
		if token.Value == "-" {
			tempMinusCount += 1
		}
	}

	// Process the correct "sign" of the first number of the postfix expression
	for i, token := range c.postfixExpr {
		if fmt.Sprintf("%T", c.postfixExpr[0].Value) == "int" {
			break
		}

		switch token.ExpressionType {
		case Symbol:
			if token.Value == "-" {
				end += 1
				minusCount += 1
			}
			if token.Value == "+" {
				end += 1
			}
		case Number:
			// Remove the first sign or signs either positive or negative from the postfixExpr
			c.postfixExpr = c.postfixExpr[end:]

			// Check for cases with only one negative sign in front: -10-12--8
			// And for cases with two negative signs in front: --10--12--8
			if minusCount == 1 {
				if fmt.Sprintf("%T", c.postfixExpr[i].Value) != "int" {
					break
				}
				c.postfixExpr[0].Value = c.postfixExpr[0].Value.(int) * -1
			}

			// Check for cases with more than 3 negatives sign in front, like: ---10--12--8
			if tempMinusCount > 1 && tempMinusCount%2 != 0 {
				c.postfixExpr[0].Value = c.postfixExpr[0].Value.(int) * -1
			}
		}
	}

	// After checking for multiple negative signs, finally start calculating the c.postfixExpr
	var a, b int
	for i, token := range c.postfixExpr {
		switch token.ExpressionType {
		case Symbol:
			if token.Value == "-" && i <= len(c.postfixExpr)-1 {
				minusCount += 1
			}
		case Number:
			c.stack = append(c.stack, Expression{Number, token.Value})
			continue
		}
		if len(c.stack) > 1 {
			// The if statements below check for various edge cases, like -10-10, 10 --10 and ---10--12--8:
			if tempMinusCount == 1 && (a == 0 && b == 0) && token.Value == "-" { // -10-10
				minusCount, tempMinusCount = 0, 0
			}

			if minusCount%2 == 0 && minusCount != 0 { // 10 -- 10
				token.Value = "+"
			}

			if tempMinusCount%2 == 1 && token.Value == "-" { // ---10--12--8
				token.Value = "+"
			}

			// Get and "pop"/remove the two last elements (numbers) from the stack
			b, a = pop(&c.stack).Value.(int), pop(&c.stack).Value.(int)

			// Perform the math operation according to the 'token' and push the result to the stack
			c.stack = append(c.stack, Expression{Number, evalSymbol(a, b, token.Value)})
			// Reset the minus counters to 0 to properly check for multiple negative signs for the next iteration
			minusCount, tempMinusCount = 0, 0
		}
	}
	// Finally return the calculated result:
	return pop(&c.stack).Value.(int)
}

// processLine is the main function that processes the input line and returns the result
func (c Calculator) processLine(line string) {
	// Since the expression is valid, proceed to append each operator and number to it:
	c.expression = c.appendValues(line)

	// After appending the values, proceed to get the postfix form of the expression:
	c.postfixExpr = c.getPostfix(c.expression)
	if c.postfixExpr != nil { // if the postfixExpr is not nil, then proceed to calculate the result
		fmt.Println(c.getTotal())
	}
	return
}

func main() {
	var c Calculator
	c.memory = make(map[string]int)

	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		line := scanner.Text()

		// Always trim/remove any leading or trailing blank spaces in the line:
		line = strings.Trim(line, " ")

		switch line {
		case "":
			continue
		case "/exit":
			fmt.Println("Bye!")
			return
		case "/help":
			fmt.Println("The program calculates the sum of numbers")
		default:
			// Check if the line is a command that begins with "/"
			if strings.HasPrefix(line, "/") {
				processCommand(line)
				continue
			}

			// Check if the line is a valid expression, if not continue and read a new input
			if !validateExpression(line) {
				continue
			}

			// If the expression is valid, then we can get the operation Type to further process the expression
			// It will be either an "Assignment" operation or a "Regular" math operation.
			c.OperationType = getOperationType(line)

			switch c.OperationType {
			case Assignment:
				c.assign(line)
				continue
			case Regular:
				c.processLine(line)
			}
		}
	}
}