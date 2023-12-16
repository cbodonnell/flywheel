class Program
{
    static void Main(string[] args)
    {
        if (args.Length == 0)
        {
            Console.WriteLine("Please specify a client type (tcp or udp)");
            return;
        }

        string clientType = args[0].ToLower();

        if (clientType == "tcp")
        {
            TCPClient tcpClient = new TCPClient();
            tcpClient.Start();
        }
        else if (clientType == "udp")
        {
            UDPClient udpClient = new UDPClient();
            udpClient.Start();
        }
        else
        {
            Console.WriteLine("Invalid client type specified");
        }
    }
}