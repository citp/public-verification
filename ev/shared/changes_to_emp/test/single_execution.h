#include <emp-tool/emp-tool.h>
#include "emp-ag2pc/emp-ag2pc.h"
using namespace std;
using namespace emp;

inline const char* hex_char_to_bin(char c) {
	switch(toupper(c)) {
		case '0': return "0000";
		case '1': return "0001";
		case '2': return "0010";
		case '3': return "0011";
		case '4': return "0100";
		case '5': return "0101";
		case '6': return "0110";
		case '7': return "0111";
		case '8': return "1000";
		case '9': return "1001";
		case 'A': return "1010";
		case 'B': return "1011";
		case 'C': return "1100";
		case 'D': return "1101";
		case 'E': return "1110";
		case 'F': return "1111";
		default: return "0";
	}
}

inline std::string hex_to_binary(std::string hex) {
	std::string bin;
	for(unsigned i = 0; i != hex.length(); ++i)
		bin += hex_char_to_bin(hex[i]);
	return bin;
}

const string circuit_file_location = macro_xstr(EMP_CIRCUIT_PATH)+string("bristol_format/");

template<typename T>
void test(int party, T* io, string name, string check_output = "", string hin = "") {
	string file = name;//circuit_file_location + name;
	BristolFormat cf(file.c_str());
	auto t1 = clock_start();
	C2PC<T> twopc(io, party, &cf);
	io->flush();
	cout << "one time:\t"<<party<<"\t" <<time_from(t1)<<endl;

	t1 = clock_start();
	twopc.function_independent();
	io->flush();
	cout << "inde:\t"<<party<<"\t"<<time_from(t1)<<endl;

	t1 = clock_start();
	twopc.function_dependent();
	io->flush();
	cout << "dep:\t"<<party<<"\t"<<time_from(t1)<<endl;

    bool TEST = true;
    bool *in; 
    bool *out;
    if (TEST) {
        in = new bool[cf.n1 + cf.n2]; //FIXME chnaged max(cf.n1, cf.n2) to cf.n1 + cf.n2
        out = new bool[cf.n3];
	    //bool *in = new bool[max(cf.n1, cf.n2)];
        if (hin.size() > 0) { //FIXME added read from in if there
            string bin = hex_to_binary(hin);
            //cout << "bin: " << endl << bin << endl;
            //cout << "bin.size(): " << bin.size() << endl;
            //cout << "cf.n1 + cf.n2: " << cf.n1 + cf.n2 << endl;
            for (int i=0; i < cf.n1 + cf.n2; ++i) {
                if (bin[i] == '0') 
                    in[i] = false;
                else if (bin[i] == '1') 
                    in[i] = true;
                else {
                    cout << "problem: " << bin[i] << endl;
                    exit(1);
                }
            }
        } else {
            memset(in, false, cf.n1 + cf.n2); //FIXME chnaged max(cf.n1, cf.n2) to cf.n1 + cf.n2
        }
    } else {
        in = new bool[max(cf.n1, cf.n2)];
        out = new bool[cf.n3];
        memset(in, false, max(cf.n1, cf.n2));
    }
    //memset(in, false, max(cf.n1, cf.n2));
	memset(out, false, cf.n3);
	t1 = clock_start();
	twopc.online(in, out, true);
	cout << "online:\t"<<party<<"\t"<<time_from(t1)<<endl;
    //FIXME ADDED
    cout << "actual output: " << endl;
    for (int i=0; i < cf.n3; ++i)
        cout << out[i];
    cout << endl;
    //FIXME END ADDED
	if(check_output.size() > 0){
		string res = "";
		for(int i = 0; i < cf.n3; ++i)
			res += (out[i]?"1":"0");
		cout << (res == hex_to_binary(check_output)? "GOOD!":"BAD!")<<endl;
	}
	delete[] in;
	delete[] out;
}


template<typename T>
void benchmark(int party, T* io, string name, string benchmark_csv_file) {
    /// CSV file of tests should have lines in the format hin,check_output\n.  hin should be a hex string of 160 characters, check_output 32.

    // read csv file of 100 tests
    ifstream csv_file(benchmark_csv_file);
    int N = 100;
    double oneTimeOffMicro[100];
    double indepOffMicro[100];
    double funcOffMicro[100];
    double onlineMicro[100];
    for (int li=0; li < N; ++li) {
        string hin;
        string check_output;
        //getline(csv_file, hin, ','); // read line up to ','
        //getline(csv_file, check_output, '\n'); // read line up to '\n'

        // old stuff
        string file = name;//circuit_file_location + name;
        BristolFormat cf(file.c_str());
        auto t1 = clock_start();
        C2PC<T> twopc(io, party, &cf);
        io->flush();
        oneTimeOffMicro[li] = time_from(t1);
        //cout << "one time:\t"<<party<<"\t" <<time_from(t1)<<endl;

        t1 = clock_start();
        twopc.function_independent();
        io->flush();
        indepOffMicro[li] = time_from(t1);
        //cout << "inde:\t"<<party<<"\t"<<time_from(t1)<<endl;

        t1 = clock_start();
        twopc.function_dependent();
        io->flush();
        funcOffMicro[li] = time_from(t1);
        //cout << "dep:\t"<<party<<"\t"<<time_from(t1)<<endl;

        bool *in; 
        bool *out;
        in = new bool[cf.n1 + cf.n2];
        out = new bool[cf.n3];
        //bool *in = new bool[max(cf.n1, cf.n2)];
        if (hin.size() > 0) {
            string bin = hex_to_binary(hin);
            //cout << "bin: " << endl << bin << endl;
            //cout << "bin.size(): " << bin.size() << endl;
            //cout << "cf.n1 + cf.n2: " << cf.n1 + cf.n2 << endl;
            for (int i=0; i < cf.n1 + cf.n2; ++i) {
                if (bin[i] == '0') 
                    in[i] = false;
                else if (bin[i] == '1') 
                    in[i] = true;
                else {
                    cout << "problem: " << bin[i] << endl;
                    exit(1);
                }
            }
        } else {
            memset(in, false, cf.n1 + cf.n2);
        }
        //memset(in, false, max(cf.n1, cf.n2));
        memset(out, false, cf.n3);
        t1 = clock_start();
        twopc.online(in, out, true);
        onlineMicro[li] = time_from(t1);
        //cout << "online:\t"<<party<<"\t"<<time_from(t1)<<endl;
        //FIXME ADDED
        //cout << "actual output: " << endl;
        for (int i=0; i < cf.n3; ++i)
            cout << out[i];
        cout << endl;
        //FIXME END ADDED
        if(check_output.size() > 0){
            string res = "";
            for(int i = 0; i < cf.n3; ++i)
                res += (out[i]?"1":"0");
            cout << (res == hex_to_binary(check_output)? "good":"BAD!")<<endl;
        }
        delete[] in;
        delete[] out;
    }
    csv_file.close();

    double avgOneTimeOffMicro = 0;
    double avgIndepOffMicro = 0;
    double avgFuncOffMicro = 0;
    double avgTotOffMicro = 0;
    double avgOnlineMicro = 0;
    for (int i=0; i < N; ++i) {
        avgOneTimeOffMicro += oneTimeOffMicro[i];
        avgIndepOffMicro += indepOffMicro[i];
        avgFuncOffMicro += funcOffMicro[i];
        avgOnlineMicro += onlineMicro[i];
    }
    avgTotOffMicro = (avgOneTimeOffMicro + avgIndepOffMicro + avgFuncOffMicro + avgOnlineMicro) / ((double)N);
    avgOneTimeOffMicro = avgOneTimeOffMicro / ((double)N);
    avgIndepOffMicro = avgIndepOffMicro / ((double)N);
    avgFuncOffMicro = avgFuncOffMicro / ((double)N);
    avgOnlineMicro = avgOnlineMicro / ((double)N);
    double varOneTimeOffMicro = 0;
    double varIndepOffMicro = 0;
    double varFuncOffMicro = 0;
    double varTotOffMicro = 0;
    double varOnlineMicro = 0;
    for (int i=0; i < N; ++i) {
        varOneTimeOffMicro += (oneTimeOffMicro[i] - avgOneTimeOffMicro) * (oneTimeOffMicro[i] - avgOneTimeOffMicro);
        varIndepOffMicro += (indepOffMicro[i] - avgIndepOffMicro) * (indepOffMicro[i] - avgIndepOffMicro);
        varFuncOffMicro += (funcOffMicro[i] - avgFuncOffMicro) * (funcOffMicro[i] - avgFuncOffMicro);
        varOnlineMicro += (onlineMicro[i] - avgOnlineMicro) * (onlineMicro[i] - avgOnlineMicro);
        varTotOffMicro += ((oneTimeOffMicro[i] + indepOffMicro[i] + funcOffMicro[i] + onlineMicro[i]) - avgTotOffMicro) *
                                ((oneTimeOffMicro[i] + indepOffMicro[i] + funcOffMicro[i] + onlineMicro[i]) - avgTotOffMicro);
    }


    cout << "P" << party << " Total offline cost average (microseconds): " << avgTotOffMicro << ", stdev: " << sqrt(varTotOffMicro) << endl;
    cout << "P" << party << "    One-time offline cost average: (microseconds): " << avgOneTimeOffMicro << ", stdev: " << sqrt(varOneTimeOffMicro) << endl;
    cout << "P" << party << "    Function-independent offline cost average: (microseconds): " << avgIndepOffMicro << ", stdev: " << sqrt(varIndepOffMicro) << endl;
    cout << "P" << party << "    Function-dependent offline cost average: (microseconds): " << avgFuncOffMicro << ", stdev: " << sqrt(varFuncOffMicro) << endl;
    cout << "P" << party << " Online cost average: (microseconds): " << avgOnlineMicro << ", stdev: " << sqrt(varOnlineMicro) << endl;
}

