# Bittorrent-in-GO

To run the seeder:

- cd seed
- go run seed.go
- in order for the seeder to seed the file it needs to be available in this directory and also the .json meta data for the file is required under seed/files directory

To run the leecher localy on the same pc:

- go to the root directory
- go run main.go <path to the torrent file> <Output filename>
  example: go run main.go debian-edu-11.6.0-amd64-netinst.iso.torrent debian-edu-11.6.0-amd64-netinst.iso

To run the leecher from a pc that is in the same LAN network(pc connected to same hotspot):

- go to /seeder/seeder.go and set the ip address of the seeding pc
  peer[0] = 127
  peer[1] = 0
  peer[2] = 0
  peer[3] = 1

  peers[0].IP = net.IP(peer)
  peers[0].Port = 8080

- cd to the root directory
- go run main.go <path to the torrent file> <Output filename>

Our implmentation only works with the debian torrent file that we have attached. Because
the seeder needs the info hash meta data(json file) in the seed/files directory, to keep track of the downloaded pieces and the actual file path.
