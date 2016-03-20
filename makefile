all:
	cd server && go build -o ../askeecs -gcflags "-N -l"
