## GoRive - Google Drive CLI

GoRive is a Google Drive client for CLI. With this application you can `list` your files with pagination and `download` file. 

---

*First Execution*

First execution of GoRive, you need to authorize your Google Drive account to GoRive. For authorization;

 1 -> go to https://developers.google.com/drive/api/v3/quickstart/go and click 'Enable the Drive API'
 
 2 -> After enable operation you need to download `credentials.json` and copy to under `$PWD/pkg/auth` folder
 
 3 -> Then you will run the app again, app will give a link for authorize. After the click and give permission operation, Google Drive generate a token. You need to copy it and paste to terminal.
 
 4 -> That's it!! You can view your drive from CLI :)

*Run GoRive*

go to `bin` folder and execute GoRive with `./gorive` command. This execution returns default 20 item. If you want to customize it just run `./gorive -count=<custom_count>`.

