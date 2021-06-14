// 首页

package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/axiaoxin-com/x-stock/core"
	"github.com/axiaoxin-com/x-stock/datacenter/eastmoney"
	"github.com/axiaoxin-com/x-stock/models"
	"github.com/axiaoxin-com/x-stock/services"
	"github.com/gin-gonic/gin"
)

// StockIndex 股票页面
func StockIndex(c *gin.Context) {
	data := gin.H{
		"PageTitle":    "X-STOCK | 股票",
		"Error":        "",
		"IndustryList": services.StockIndustryList,
	}
	c.HTML(http.StatusOK, "stock_index.html", data)
	return
}

// ParamStockSelector StockSelector 请求参数
type ParamStockSelector struct {
	Filter            eastmoney.Filter
	CheckerOptions    core.CheckerOptions
	FilterWithChecker bool `form:"selector_with_checker"`
}

// StockSelector 返回基本面筛选结果json
func StockSelector(c *gin.Context) {
	data := gin.H{
		"PageTitle": "X-STOCK | 股票 | 基本面筛选",
		"Error":     "",
		"Stocks":    models.StockList{},
	}

	param := ParamStockSelector{}
	if err := c.ShouldBind(&param); err != nil {
		data["Error"] = err.Error()
		c.JSON(http.StatusOK, data)
		return
	}
	var checker core.Checker
	if param.FilterWithChecker {
		checker = core.NewChecker(c, param.CheckerOptions)
	}

	selector := core.NewSelector(c, param.Filter, &checker)
	stocks, err := selector.AutoFilterStocks(c)
	if err != nil {
		data["Error"] = err.Error()
		c.JSON(http.StatusOK, data)
		return
	}
	dlist := models.ExportorDataList{}
	for _, s := range stocks {
		dlist = append(dlist, models.NewExportorData(c, s))
	}
	data["Stocks"] = dlist
	c.JSON(http.StatusOK, data)
	return
}

// ParamStockChecker StockChecker 请求参数
type ParamStockChecker struct {
	Keyword        string `form:"checker_keyword"`
	CheckerOptions core.CheckerOptions
}

// StockChecker 处理个股检测
func StockChecker(c *gin.Context) {
	data := gin.H{
		"PageTitle": "X-STOCK | 股票 | 个股检测",
		"Error":     "",
	}
	param := ParamStockChecker{}
	if err := c.ShouldBind(&param); err != nil {
		data["Error"] = err.Error()
		c.JSON(http.StatusOK, data)
		return
	}
	if param.Keyword == "" {
		data["Error"] = "请填写股票代码或简称"
		c.JSON(http.StatusOK, data)
		return
	}
	searcher := core.NewSearcher(c)
	keywords := strings.Split(param.Keyword, "/")
	stocks, err := searcher.Search(c, keywords)
	if err != nil {
		data["Error"] = err.Error()
		c.JSON(http.StatusOK, data)
		return
	}
	checker := core.NewChecker(c, param.CheckerOptions)
	results := []core.CheckResult{}
	names := []string{}
	for _, stock := range stocks {
		result, _ := checker.CheckFundamentals(c, stock)
		results = append(results, result)
		name := fmt.Sprintf("%s-%s", stock.BaseInfo.SecurityNameAbbr, stock.BaseInfo.Secucode)
		names = append(names, name)
	}
	data["Results"] = results
	data["Names"] = names
	c.JSON(http.StatusOK, data)
	return
}