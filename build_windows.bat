mkdir bin
cd l2 && go build -o ..\bin\luna-l2.exe luna_l2.go && cd ..\
cd l2ld && go build -o ..\bin\l2ld.exe l2ld.go && cd ..\
cd las && go build -o ..\bin\las.exe las.go && cd ..\
cd lcc && go build -o ..\bin\lcc.exe lcc.go && cd ..\
cd lcc1 && go build -o ..\bin\lcc1.exe lcc1.go && cd ..\
type nul > ..\bin\libs.conf
