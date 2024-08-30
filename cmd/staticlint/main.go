//	Пакет main поддерживает следующие анализаторы
//
// exitanalyzer - самописный анализатор os.Exit,
// errcheck - проверка на обработку ошибок,
// analysis - стандартный пакет анализаторов.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"github.com/kisielk/errcheck/errcheck"
	"honnef.co/go/tools/staticcheck" //staticcheck.io

	osexit "github.com/SversusN/shortener/cmd/staticlint/osexit"
)

func main() {
	var chks []*analysis.Analyzer

	// Добаляем все анализаторы staticcheck.
	for _, v := range staticcheck.Analyzers {
		chks = append(chks, v.Analyzer)
	}

	// Дополнительные анализаторы
	chks = append(
		chks,
		osexit.OSExitAnalyzer, // Проверяем os.Exit в main.
		printf.Analyzer,       // Проверяем формтированную печать printf.
		shadow.Analyzer,       // Проверяем shadow-переопределения.
		structtag.Analyzer,    // Проверяем правильность типов структур.
		errcheck.Analyzer,     // Проверяем обработку ошибок.
	)

	//Запускаем multichecker
	multichecker.Main(chks...)
}
