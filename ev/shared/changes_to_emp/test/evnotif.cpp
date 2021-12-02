#include <emp-tool/emp-tool.h>
#include "test/single_execution.h"
using namespace std;
using namespace emp;

int main(int argc, char** argv) {
	int party, port;
	parse_party_and_port(argv, &party, &port);
	NetIO* io = new NetIO(party==ALICE ? nullptr:IP, port);
//	io->set_nodelay();
    //string input00 = "0000000000000000000000000000000000000000000000000000000000000000";
    //string inputm0k1 = "00000000000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF";
    //string aes00 = "66E94BD4EF8A2C3B884CFA59CA342B2E";
    string aesm0k1 = "A1F6258C877D5FCD8964484538BFC92C";

    // input00010 should have ika = adkey = ikb = nonce = 0, ek = -1.  Output should be same as AES(m=0, k=-1)
    string input00010 = string(32, '0') + string(32, '0') + string(32, '0') + string(32, 'F') + string(32, '0');

    // input10110 should have ika = AA...A, adkey = 0, ikb = AA...A, ek = -1, nonce = 0
    string input10110 = string(32, 'A') + string(32, '0') + string(32, 'A') + string(32, 'F') + string(32, '0');

    // inputABABF should have ika = AA...A, adkey = BB...B, ikb = AA...A, ek = -1, nonce = BB...B
    string inputABABF = string(32, 'A') + string(32, 'B') + string(32, 'A') + string(32, 'F') + string(32, 'B');

	//test<NetIO>(party, io, circuit_file_location+"AES-non-expanded.txt", aes00, input00);
    //cout << "expected output: " << endl << hex_to_binary(aes00) << endl;

    // ika   = 21DA22CF8FA98B69AB4DFF65AB0850F5
    // adkey = 3A1F3B78D29A319B682EAC046750CAB4
    // ikb   = 21DA22CF8FA98B69AB4DFF65AB0850F5
    // ek    = D4A48FC5834D6F1C62A04F8CAFBEFB19
    // nonce = DE45C251A6A6FD91C29191A48D45CAB8
    // out   = 860A6A39C34DA88589016F3DE950C4D2
    string randin = "21DA22CF8FA98B69AB4DFF65AB0850F53A1F3B78D29A319B682EAC046750CAB421DA22CF8FA98B69AB4DFF65AB0850F5D4A48FC5834D6F1C62A04F8CAFBEFB19DE45C251A6A6FD91C29191A48D45CAB8";
    string randout = "860A6A39C34DA88589016F3DE950C4D2";
    
	//test<NetIO>(party, io, circuit_file_location+"AES-non-expanded.txt", aesm0k1, inputm0k1);
	//test<NetIO>(party, io, "/root/shared/ev_notif_aes.txt", aesm0k1, input00010);
	//test<NetIO>(party, io, "/root/shared/ev_notif_aes.txt", aesm0k1, input10110);
	//test<NetIO>(party, io, "/root/shared/ev_notif_aes.txt", aesm0k1, inputABABF);
	test<NetIO>(party, io, "/root/shared/ev_notif_aes.txt", randout, randin);
    cout << "expected output: " << endl << hex_to_binary(randout) << endl;
	delete io;
	return 0;
}
