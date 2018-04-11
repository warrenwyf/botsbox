package analyzers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

var html = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html lang="zh" xml:lang="zh" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta content="text/html; charset=utf-8" http-equiv="content-type" />
  <title>微软 Bing 搜索 - 国内版</title>
  <link rel="icon" sizes="any" mask="" href="/fd/s/a/hp/bingcn.svg" />
  <meta name="theme-color" content="#4F4F4F" />
  <link href="/sa/simg/bing_p_rr_teal_min.ico" rel="shortcut icon" />
  <meta content="微软Bing搜索是国际领先的搜索引擎，为中国用户提供网页、图片、视频、词典、翻译、地图等全球信息搜索服务。" name="description" />
  <meta content="NOODP" name="ROBOTS" />
  <meta name="pinterest" content="nohover" />
 </head>
 <body class="zhs zh-CN wkit" onload="hpResize(_ge('bgDiv'));_ge('sb_form_q').focus();if(_w.lb)lb();" onfocus="">
  <script type="text/javascript">//<![CDATA[
_G.AppVer="8_01_0_000000"; var _H={}; _H.mkt = "zh-CN";_H.trueMkt = "zh-CN";_H.imgName = "CardonCactus";;_H.sbe = 1; var g_vidOn=0; var g_hasVid=0; var g_IMVL = 0; var g_NPLE =0;var g_hptse = 1; var g_crsInst =0; _H.rbc =0;_H.clientLog = 1;_H.startTm = _w.performance && _w.performance.timing ? _w.performance.timing.responseStart : null; _H.crsEL =0;_H.focusHideCrs =0; _H.plf = {};_H.hpqs = 1;_H.fullimg = 1;_H.hpaimg = 1;;var sj_b=_d.body;_H.plf.officeMenu = 1;;var Identity; (function(Identity) { Identity.sglid =false; Identity.orgIdPhotoUrl ="https://business.bing.com/api/v2/search/download?DocumentType=ContactPhoto\u0026Id={0}"; })(Identity || (Identity = {}));;BM.trigger();
//]]></script>
  <table id="hp_table">
   <tbody>
    <tr>
     <td id="hp_cellCenter" class="hp_hd">
      <div id="hp_container">
       <div id="bgDiv" data-minhdhor="" data-minhdver="" data-priority="0">
        <div id="hp_vidwrp"></div>
        <video id="vid" onended="_w.VM &amp;&amp; VM.pause();" autobuffer="autobuffer" preload="auto" oncontextmenu="return false"></video>
       </div>
       <div id="sbox" class="sw_sform" data-priority="0">
        <div class="hp_sw_logo hpcLogoWhite" data-fixedpriority="0">
         必应
        </div>
        <div id="est_switch">
         <div id="est_cn" class="est_selected">
          国内版
         </div>
         <div id="est_en" class="est_unselected">
          国际版
         </div>
        </div>
        <div class="search_controls">
         <a id="hpinsthk" aria-hidden="true" tabindex="-1" href="javascript:void(0)" h="ID=SERP,5092.1"><span></span></a>
         <form action="/search" onsubmit="var id = _ge('hpinsthk').getAttribute('h'); return si_T(id);" id="sb_form" class="sw_box">
          <div class="b_searchboxForm" role="search">
           <input class="b_searchbox" id="sb_form_q" name="q" title="输入搜索词" type="search" value="" maxlength="100" autocapitalize="off" autocorrect="off" autocomplete="off" spellcheck="false" />
           <input type="submit" class="b_searchboxSubmit" id="sb_form_go" title="搜索" tabindex="0" name="go" />
           <input id="sa_qs" name="qs" value="ds" type="hidden" />
           <input type="hidden" value="QBLH" name="form" />
          </div>
         </form>
        </div>
       </div>
       <div class="shader_left"></div>
       <div class="shader_right"></div>
       <div id="hp_sw_hdr" class="hp_hor_hdr" data-priority="1">
        <div class="sw_tb">
         <ul id="sc_hdu" class="sc_hl1 hp_head_nav" role="navigation">
          <li id="scpt0" class=""><a id="scpl0" aria-owns="scpc0" aria-controls="scpc0" aria-expanded="false" onclick="hpulc4hdr();selectScope(this, 'images');" href="/images?FORM=Z9LH" h="ID=SERP,5024.1">图片</a>
           <div id="scpc0" role="group" aria-labelledby="scpl0" aria-hidden="true" aria-expanded="false"></div></li>
          <li id="scpt1" class=""><a id="scpl1" aria-owns="scpc1" aria-controls="scpc1" aria-expanded="false" onclick="hpulc4hdr();selectScope(this, 'video');" href="/videos?FORM=Z9LH1" h="ID=SERP,5025.1">视频</a>
           <div id="scpc1" role="group" aria-labelledby="scpl1" aria-hidden="true" aria-expanded="false"></div></li>
          <li id="scpt2" class=""><a id="scpl2" aria-owns="scpc2" aria-controls="scpc2" aria-expanded="false" onclick="hpulc4hdr();selectScope(this, 'academic');" href="/academic/?FORM=Z9LH2" h="ID=SERP,5026.1">学术</a>
           <div id="scpc2" role="group" aria-labelledby="scpl2" aria-hidden="true" aria-expanded="false"></div></li>
          <li id="scpt3" class=""><a id="scpl3" aria-owns="scpc3" aria-controls="scpc3" aria-expanded="false" onclick="hpulc4hdr();selectScope(this, 'dictionary');" href="/dict?FORM=Z9LH3" h="ID=SERP,5027.1">词典</a>
           <div id="scpc3" role="group" aria-labelledby="scpl3" aria-hidden="true" aria-expanded="false"></div></li>
          <li id="scpt4" class=""><a id="scpl4" aria-owns="scpc4" aria-controls="scpc4" aria-expanded="false" onclick="hpulc4hdr();selectScope(this, 'local');" href="/maps?FORM=Z9LH4" h="ID=SERP,5028.1">地图</a>
           <div id="scpc4" role="group" aria-labelledby="scpl4" aria-hidden="true" aria-expanded="false"></div></li>
          <li id="hdr_spl">|</li>
          <li id="office"><a id="off_link" aria-owns="off_menu_cont" aria-controls="off_menu_cont" aria-expanded="false" target="_blank" onclick="hpulc4hdr();" href="http://www.office.com?WT.mc_id=O16_BingHP" h="ID=SERP,5015.1">Office Online</a>
           <div id="off_menu_cont" aria-labelledby="off_link" aria-expanded="false" aria-hidden="true" class="sc_pc" data-officemenuroot="office">
            <ul class="om">
             <li><a id="officemenu_word" title="Word Online" target="_blank" onclick="hpulc4hdr();" href="https://office.live.com/start/Word.aspx?WT.mc_id=O16_BingHP" h="ID=SERP,5029.1">
               <div>
                <div class="oml_img" id="officemenu_word_img"></div>
                <div class="itm_desc">
                 Word Online
                </div>
               </div></a></li>
             <li><a id="officemenu_excel" title="Excel Online" target="_blank" onclick="hpulc4hdr();" href="https://office.live.com/start/Excel.aspx?WT.mc_id=O16_BingHP" h="ID=SERP,5030.1">
               <div>
                <div class="oml_img" id="officemenu_excel_img"></div>
                <div class="itm_desc">
                 Excel Online
                </div>
               </div></a></li>
             <li><a id="officemenu_powerpoint" title="PowerPoint Online" target="_blank" onclick="hpulc4hdr();" href="https://office.live.com/start/PowerPoint.aspx?WT.mc_id=O16_BingHP" h="ID=SERP,5031.1">
               <div>
                <div class="oml_img" id="officemenu_powerpoint_img"></div>
                <div class="itm_desc">
                 PowerPoint Online
                </div>
               </div></a></li>
             <li><a id="officemenu_onenote" title="OneNote Online" target="_blank" onclick="hpulc4hdr();" href="https://www.onenote.com/notebooks?WT.mc_id=O16_BingHP" h="ID=SERP,5032.1">
               <div>
                <div class="oml_img" id="officemenu_onenote_img"></div>
                <div class="itm_desc">
                 OneNote Online
                </div>
               </div></a></li>
             <li><a id="officemenu_sway" title="Sway" target="_blank" onclick="hpulc4hdr();" href="https://www.sway.com?WT.mc_id=O16_BingHP&amp;utm_source=O16Bing&amp;utm_medium=Nav&amp;utm_campaign=HP" h="ID=SERP,5033.1">
               <div>
                <div class="oml_img" id="officemenu_sway_img"></div>
                <div class="itm_desc">
                 Sway
                </div>
               </div></a></li>
             <li><a id="officemenu_docscom" title="Docs.com" target="_blank" onclick="hpulc4hdr();" href="https://www.docs.com?WT.mc_id=O16_BingHP" h="ID=SERP,5034.1">
               <div>
                <div class="oml_img" id="officemenu_docscom_img"></div>
                <div class="itm_desc">
                 Docs.com
                </div>
               </div></a></li>
             <li><a id="officemenu_onedrive" title="OneDrive" target="_blank" onclick="hpulc4hdr();" href="https://onedrive.live.com/?gologin=1&amp;WT.mc_id=O16_BingHP" h="ID=SERP,5035.1">
               <div>
                <div class="oml_img" id="officemenu_onedrive_img"></div>
                <div class="itm_desc">
                 OneDrive
                </div>
               </div></a></li>
             <li><a id="officemenu_calendar" title="日历" target="_blank" onclick="hpulc4hdr();" href="https://calendar.live.com/?WT.mc_id=O16_BingHP" h="ID=SERP,5036.1">
               <div>
                <div class="oml_img" id="officemenu_calendar_img"></div>
                <div class="itm_desc">
                 日历
                </div>
               </div></a></li>
             <li><a id="officemenu_people" title="人脉" target="_blank" onclick="hpulc4hdr();" href="https://outlook.live.com/owa/?path=/people&amp;WT.mc_id=O16_BingHP" h="ID=SERP,5037.1">
               <div>
                <div class="oml_img" id="officemenu_people_img"></div>
                <div class="itm_desc">
                 人脉
                </div>
               </div></a></li>
            </ul>
           </div></li>
          <li id="outlook"><a id="off_link" aria-owns="off_menu_cont" aria-controls="off_menu_cont" aria-expanded="false" target="_blank" onclick="hpulc4hdr();" href="https://outlook.com/?WT.mc_id=O16_BingHP?mkt=zh-CN" h="ID=SERP,5016.1">Outlook.com</a></li>
         </ul>
         <div id="hp_id_hdr">
          <div id="id_h" role="complementary" aria-label="帐户奖励和偏好设置" data-priority="2">
           <a id="id_l" class="id_button" role="button" aria-haspopup="true" href="javascript:void(0);" h="ID=SERP,5061.1"><span id="id_s" aria-hidden="false">登录</span><span class="sw_spd id_avatar" id="id_a" aria-hidden="false" aria-label="默认个人资料图片"></span><span id="id_n" style="display:none" aria-hidden="true"></span><img id="id_p" class="id_avatar sw_spd" style="display:none" aria-hidden="true" src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAEALAAAAAABAAEAAAIBTAA7" aria-label="个人资料图片" onerror="FallBackToDefaultProfilePic(this)" /></a>
           <span id="id_d" _iid="SERP.5069"></span>
           <a id="id_sc" class="idp_ham hphbtop" aria-label="首选项" aria-expanded="false" aria-controls="id_hbfo" role="button" tabindex="0" href="javascript:void(0);" h="ID=SERP,5067.1"></a>
           <span id="id_hbfo" _iid="SERP.5068" class="slide_up hpfo hb_hpqexp" aria-hidden="true" aria-haspopup="true" role="menu"></span>
          </div>
         </div>
        </div>
       </div>
       <div id="lap_w" data-ajaxiid="5043" data-date="20180404"></div>
       <div id="hp_bottomCell">
        <div id="hp_ctrls" class=" cnhpCtrls cnlifeaa" data-tbarhidden="">
         <div id="sh_rdiv" data-priority="1">
          <a id="sh_shqzl" title="QQ空间" target="_blank" href="http://sns.qzone.qq.com/cgi-bin/qzshare/cgi_qzshare_onekey?title={0}&amp;summary={1}&amp;url={2}&amp;pics={3}" h="ID=SERP,5042.1">
           <div id="sh_shqz" class="hpcsQzone sh_hide"></div></a>
          <a id="sh_shwbl" title="微博" target="_blank" href="http://service.weibo.com/share/share.php?title={1}&amp;placeholder={0}&amp;url={2}&amp;pic={3}" h="ID=SERP,5041.1">
           <div id="sh_shwb" class="hpcsWeibo sh_hide"></div></a>
          <a id="sh_shwcl" title="微信" href="javascript:void(0)" h="ID=SERP,5040.1">
           <div id="sh_shwc" class="hpcsWechat sh_hide"></div></a>
          <a id="sh_shl" class="sc_lightdis" title="分享" data-sharedcountenabled="True" href="javascript:void(0)" h="ID=SERP,5039.1">
           <div id="sh_sh" class="hpcShare"></div></a>
          <div id="sh_shwcp" class="sh_hide">
           <div id="sh_shwcpq">
            <img id="sh_shwci0" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180404" />
            <img id="sh_shwci1" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180403" />
            <img id="sh_shwci2" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180402" />
            <img id="sh_shwci3" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180401" />
            <img id="sh_shwci4" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180331" />
            <img id="sh_shwci5" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180330" />
            <img id="sh_shwci6" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180329" />
            <img id="sh_shwci7" class="sh_hide" data-pageurl="http://cn.bing.com/coverstory?ensearch=0%26date=20180328" />
           </div>
          </div>
          <div id="showBingAppQR" class="bingAppQRHide">
           <div id="downloadBingAppTip">
            下载手机必应
           </div>
           <div id="shwBingAppQR">
            <img class="rms_img" src="/rms/rms%20answers%20Homepage%20ZhCn$BingAppQR/ic/e19fde95/ce98ce99.png" />
           </div>
           <div class="bingAppQRVLine"></div>
          </div>
          <a title="扫一扫下载手机必应" href="javascript:void(0)" h="ID=SERP,5038.1">
           <div id="shBingAppQR"></div></a>
          <a role="button" id="sh_psv" class="sh_psl" title="暂停视频" aria-label="暂停视频" tabindex="-1" aria-hidden="true" href="javascript:void(0)" h="ID=SERP,5047.1">
           <div id="sh_ps" class="hpcPause"></div></a>
          <a role="button" class="sh_pll" title="播放视频" aria-label="播放视频" tabindex="-1" aria-hidden="true" href="javascript:void(0)" h="ID=SERP,5046.1">
           <div id="sh_pl" class="hpcPlay"></div></a>
          <a role="button" id="sh_igl" title="上一个图像" aria-label="上一个图像" href="?FORM=HYLH#" h="ID=SERP,5058.1">
           <div class="sc_lightdis">
            <div id="sh_lt" class="hpcPrevious"></div>
           </div></a>
          <a role="button" id="sh_igr" title="下一个图像" aria-label="下一个图像" href="?FORM=HYLH1#" h="ID=SERP,5057.1">
           <div class="sc_lightdis">
            <div id="sh_rt" class="hpcNext"></div>
           </div></a>
          <a id="sh_cp" class="sc_light" title="武伦柱 (&copy; Ed Reardon/Alamy)" aria-label="主页图片信息" role="button" target="_blank" href="javascript:void(0)" h="ID=SERP,5038.2">
           <div>
            <div id="sh_cp_in" class="hpcCopyInfo"></div>
           </div></a>
         </div>
        </div>
        <div>
         <div id="hp_notf" role="contentinfo"></div>
         <div id="hp_tbar" class=" hp_cnCarousel"></div>
        </div>
        <footer id="b_footer" class="b_footer" role="contentinfo" aria-label="页脚" data-priority="0">
         <div id="b_footerItems">
          <span>&copy; 2018 Microsoft</span>
          <ul>
           <li><span>京ICP备10036305号</span></li>
           <li><span>京公网安备11010802022657号</span></li>
           <li><a id="sb_privacy" href="http://go.microsoft.com/fwlink/?LinkId=521839&amp;CLCID=804" h="ID=SERP,5074.1">隐私声明和 Cookie</a></li>
           <li><a id="sb_legal" href="http://go.microsoft.com/fwlink/?LinkID=246338&amp;CLCID=804" h="ID=SERP,5075.1">法律声明</a></li>
           <li><a id="sb_advertise" href="http://go.microsoft.com/?linkid=9844344" h="ID=SERP,5076.1">广告</a></li>
           <li><a id="sb_report" href="http://go.microsoft.com/fwlink/?LinkID=275671&amp;clcid=0x04" h="ID=SERP,5077.1">报告</a></li>
           <li><a id="sb_help" target="_blank" href="http://go.microsoft.com/fwlink/?LinkID=617297&amp;clcid=0x04" h="ID=SERP,5078.1">帮助</a></li>
           <li><a id="sb_feedback" href="#" h="ID=SERP,5079.1">反馈</a></li>
          </ul>
         </div>
         <!--foo-->
        </footer>
       </div>
      </div></td>
    </tr>
   </tbody>
  </table>
 </body>
</html>
`

func Test_actOnHtmlUrl(t *testing.T) {
	var url string

	url = actOnHtmlUrl("http://xyz.com/foo/bar", nil, "https://abc.com/foo/bar.html")
	if url != "http://xyz.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnHtmlUrl("//xyz.com/foo/bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://xyz.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnHtmlUrl("./bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://abc.com/foo/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}

	url = actOnHtmlUrl("/bar", nil, "https://abc.com/foo/bar.html")
	if url != "https://abc.com/bar" {
		t.Fatalf(`Wrong URL result: "%s"`, url)
	}
}

func Test_extractHtmlValue(t *testing.T) {
	analyzer := NewHtmlAnalyzer(nil)

	r := bytes.NewReader(([]byte)(html))

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatal("Parse HTML error:", err)
	}

	est_cn := analyzer.extractHtmlValue(doc.Selection, "$[#est_cn].$text")
	if strings.TrimSpace(est_cn) != "国内版" {
		t.Fatal("extractHtmlValue() get wrong result")
	}
}
