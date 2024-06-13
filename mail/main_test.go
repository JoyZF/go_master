package mail

import "testing"

func TestSendEmail(t *testing.T) {
	body := `<center><div style="width:750px;height:20px;background:black;text-align:center;line-height:20px;font-size:12px;"><a href="http://game.163.com/mail/2022/1019/1/" target="_blank" style="color:white;">如果您看不到这封邮件，请点击这里查看</a></div><div style="margin:0 auto;width:750px;height:7642px;background:url(https://nie.res.netease.com/nie/mail/6dECXFPQnZ.jpg) no-repeat;overflow:hidden;"><table cellspacing="0" border="0" cellpadding="0" style="width:750px;border-collapse:collapse;"><tr style="height:1196px"><td><p style="display: block;font-family: 微软雅黑,simsun,serif;word-break: break-all;line-height: 40px;margin: 0 auto;padding: 0;width: 62%;height: 40px;font-size: 36px;text-align: center;margin-left: 19%;margin-top: 92%;color: #e5c459;">${code}</p></td></tr><tr style="height:48px"><td style="height:48px;vertical-align:top;"><a href="https://projectextreme.neteasegames.com/" title="" target="_blank" style="display:block;width:478px;height:48px;border:0;vertical-align:top;text-align:left;overflow:hidden;float:left;margin-left:136px"><img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAALSURBVBhXY2AAAgAABQABqtXIUQAAAABJRU5ErkJggg==" style="width:100%;height:100%;border:0;opacity: 0;" /></a></td></tr></table></div></center>`
	SendEmail("test-eocuv0ot4@srv1.mail-tester.com", "FakePunk Closed Alpha Invitation", body)
}
