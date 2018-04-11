package analyzers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/beevik/etree"

	"../rule"
	"../target"
)

var xml = `
<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" version="2.0" xmlns:media="http://search.yahoo.com/mrss/">
<channel><title><![CDATA[test_title]]></title><description><![CDATA[一起探索地理位置的价值]]></description><link>http://blog.geohey.com/</link><generator>Ghost 0.6</generator><lastBuildDate>Sun, 01 Apr 2018 11:57:45 GMT</lastBuildDate><atom:link href="http://blog.geohey.com/rss/" rel="self" type="application/rss+xml"/><ttl>60</ttl>
<item><title><![CDATA[成都，成都，时隔半年多，房价怎么说]]></title><description><![CDATA[<p>可...可恶啊，2018年已经过了一个季度了，这一年定的小目标完成多少了？</p>

<p>看书？运动？减肥？学习？ 谈恋爱？ 毕业去向？ 涨工资？</p>

<p>人艰不拆啊人艰不拆，很多人新年的小目标可能到现在都忘了...౿(།﹏།)૭</p>

<p>倒是笔者过年回家的时候，很多伙伴都说要去成都发展，或者已经就在成都工作了。</p>

<p>可不，有一句话不是说：“成都，一座你来了不想走的城市”。
<img src="http://blog.geohey.com/content/images/2018/04/timg.jpg" alt=""></p>

<p>那可是很有道理的，毕竟是天府之国，气候湿润，吃辣养生的地方，美女多什么都就不谈了，西南新一线城市就不谈了...</p>

<p>那我们谈什么呢？</p>

<p>既然在这座城不想走，那就得安居乐业，所谓安居乐业，首先要安居，然后才能乐业。</p>

<p>所以咱们来看看房价的情况，给大家参考参考。</p>

<p>笔者整理了部分成都2017年8月时候的楼盘价，然后和2018年3月的楼盘价进行了对比（相隔大半年），总计有5000+楼盘的房价变化，供各位看官(仅供)参考参考。</p>

<p><center> <strong>（一）总览</strong></center> <br>
2017年8月 成都全市平均房价为 <strong>9702</strong>元/㎡</p>

<p>2018年3月 成都全市平均房价为</p>]]></description><link>http://blog.geohey.com/cheng-du-cheng-du-shi-ge-ban-nian-duo-fang-jie-zen-yao-shuo/</link><guid isPermaLink="false">22165b86-a620-4288-8798-b9770aec0bbc</guid><dc:creator><![CDATA[冬雨Zz]]></dc:creator><pubDate>Sun, 01 Apr 2018 11:51:32 GMT</pubDate><media:content url="http://blog.geohey.com/content/images/2018/04/p1.jpg" medium="image"/><content:encoded><![CDATA[<img src="http://blog.geohey.com/content/images/2018/04/p1.jpg" alt="成都，成都，时隔半年多，房价怎么说"><p>可...可恶啊，2018年已经过了一个季度了，这一年定的小目标完成多少了？</p>

<p>看书？运动？减肥？学习？ 谈恋爱？ 毕业去向？ 涨工资？</p>

<p>人艰不拆啊人艰不拆，很多人新年的小目标可能到现在都忘了...౿(།﹏།)૭</p>

<p>倒是笔者过年回家的时候，很多伙伴都说要去成都发展，或者已经就在成都工作了。</p>

<p>可不，有一句话不是说：“成都，一座你来了不想走的城市”。
<img src="http://blog.geohey.com/content/images/2018/04/timg.jpg" alt="成都，成都，时隔半年多，房价怎么说"></p>

<p>那可是很有道理的，毕竟是天府之国，气候湿润，吃辣养生的地方，美女多什么都就不谈了，西南新一线城市就不谈了...</p>

<p>那我们谈什么呢？</p>

<p>既然在这座城不想走，那就得安居乐业，所谓安居乐业，首先要安居，然后才能乐业。</p>

<p>所以咱们来看看房价的情况，给大家参考参考。</p>

<p>笔者整理了部分成都2017年8月时候的楼盘价，然后和2018年3月的楼盘价进行了对比（相隔大半年），总计有5000+楼盘的房价变化，供各位看官(仅供)参考参考。</p>

<p>所以总的说来，成都依然是一个超级抢手的热门城市(毕竟要当西南第一城市)，关于房价上涨的原因，我们就不讨论了，不是一篇文章了能解释清楚的，所以呢给各位伙伴一些数据参考参考。</p>

<p>预祝想在成都定居的各位伙伴早日能有自己温馨的窝窝，吃辣不再长痘痘。</p>]]></content:encoded>
</item>
<item><title><![CDATA[这是第二个条目]]></title>
</item>
</channel>
</rss>
`

func Test_XmlAnalyzer_ParseBytes(t *testing.T) {
	r := rule.NewRule()
	analyzer := NewXmlAnalyzer(r)

	baseTarget := target.NewTarget()

	output := rule.NewListOutput()
	output.Name = "rss_dataset"
	output.Selector = "//item"
	output.Data["raw"] = "$raw"
	baseTarget.ListOutputs = []*rule.ListOutput{output}

	result, err := analyzer.ParseBytes(([]byte)(xml), "", baseTarget)
	if err != nil {
		t.Fatal("ParseBytes error:", err)
	}

	if len(result.SinkPacks) != 2 {
		t.Fatal("Parse SinkPacks get wrong result:", result.SinkPacks)
	}
}

func Test_extractXmlValue(t *testing.T) {
	analyzer := NewXmlAnalyzer(nil)

	analyzer.raw = ([]byte)(xml)

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(analyzer.raw); err != nil {
		t.Fatal("Parse XML error", err)
	}

	analyzer.doc = doc

	raw := strings.TrimSpace(analyzer.extractXmlValue(doc.Root(), "$raw"))
	if !strings.HasPrefix(raw, "<?xml") || !strings.HasSuffix(raw, "</rss>") {
		t.Fatal("extractXmlValue() get wrong result", raw)
	}

	title := analyzer.extractXmlValue(doc.Root(), "$[//title].$text")
	if title != "test_title" {
		t.Fatal("extractXmlValue() get wrong result", title)
	}
}

func Test_cleanXmlCharacter(t *testing.T) {
	reader := bytes.NewReader([]byte{0x0008, 0x0061})
	b := cleanXmlCharacter(reader)
	if string(b) != "a" {
		t.Fatal("cleanXmlCharacter() get wrong result", b)
	}
}
