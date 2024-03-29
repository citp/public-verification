#include <emp-tool/emp-tool.h>
#include "test/single_execution.h"
using namespace std;
using namespace emp;

int main(int argc, char** argv) {
	int party, port;
	parse_party_and_port(argv, &party, &port);
	NetIO* io = new NetIO(party==ALICE ? nullptr:IP, port);
	//io->set_nodelay();
	//test(party, io, "test/ands.txt");
	benchmark(party, io, "/root/shared/ev_notif_aes.txt", "/root/shared/bench_tests.csv");
	delete io;
	return 0;
}
