package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jszwec/csvutil"
	"github.com/sclevine/agouti"
	"github.com/yukpiz/monthly_automation/config"
)

var (
	cfgPath = flag.String("config", "", "config file path")
)

func main() {
	flag.Parse()
	if len(*cfgPath) == 0 {
		fmt.Println("empty config file path")
		os.Exit(1)
	}
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	if err := StartCrawl(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

}

func StartCrawl(cfg *config.Config) error {
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=500,500",
		}), agouti.Debug,
	)
	if err := driver.Start(); err != nil {
		return err
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		return err
	}

	if err := page.Navigate(cfg.RecoruConfig.LoginURL); err != nil {
		return err
	}
	Sleep(1)

	cele := page.FindByName("contractId")
	if err := cele.Fill(cfg.RecoruConfig.ContractID); err != nil {
		return err
	}
	lele := page.FindByID("authId")
	if err := lele.Fill(cfg.RecoruConfig.LoginID); err != nil {
		return err
	}
	pele := page.FindByID("password")
	if err := pele.Fill(cfg.RecoruConfig.LoginPassword); err != nil {
		return err
	}

	lbtn := page.FindByButton("ログイン")
	if err := lbtn.Submit(); err != nil {
		return err
	}
	Sleep(1)

	html, err := page.HTML()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return err
	}
	tbl := doc.Find("#ID-attendanceChartGadgetTable")
	rows := tbl.Find("tbody > tr")

	type Data struct {
		WeekDay        string `csv:"week_day"`
		WorkStartTime  string `csv:"start"`
		WorkEndTime    string `csv:"end"`
		BreakTimeRange string `csv:"break"`
		WorkMemo       string `csv:"memo"`
	}

	var datas []*Data
	rows.Each(func(i int, tr *goquery.Selection) {
		data := &Data{}
		tr.Find("td.item-day > label").Each(func(i int, lbl *goquery.Selection) {
			fmt.Println(lbl.Text())
			pat := regexp.MustCompile(`(\d+)\/(\d+).*`)
			m := pat.FindAllStringSubmatch(strings.TrimSpace(lbl.Text()), -1)
			if len(m) != 1 || len(m[0]) != 3 {
				fmt.Println("not pattern match")
				return
			}

			mo, err := strconv.Atoi(m[0][1])
			if err != nil {
				fmt.Printf("%+v\n", err)
				return
			}

			d, err := strconv.Atoi(m[0][2])
			if err != nil {
				fmt.Printf("%+v\n", err)
				return
			}

			now := time.Now()
			data.WeekDay = fmt.Sprintf("%d/%d/%d", mo, d, now.Year())
		})

		tr.Find("td.item-worktimeStart").Each(func(i int, td *goquery.Selection) {
			data.WorkStartTime = strings.TrimSpace(td.Text())
		})

		tr.Find("td.item-worktimeEnd").Each(func(i int, td *goquery.Selection) {
			data.WorkEndTime = strings.TrimSpace(td.Text())
			td = td.Next()
			data.BreakTimeRange = strings.TrimSpace(td.Text())
		})

		tr.Find("td.item-worktimeMemo").Each(func(i int, td *goquery.Selection) {
			data.WorkMemo = fmt.Sprintf("\"%s\"", strings.TrimSpace(td.Text()))
		})

		if len(data.WorkStartTime) == 0 {
			return
		}
		datas = append(datas, data)
	})

	b, err := csvutil.Marshal(datas)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", string(b))

	return nil
}

func Sleep(sec int) {
	fmt.Printf("Wait for %d seconds...\n", sec)
	time.Sleep(time.Duration(sec) * time.Second)
}
