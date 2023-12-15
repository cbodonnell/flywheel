using System;
using System.Net.Sockets;
using System.Text;

class Program
{
    static void Main()
    {
        TcpClient tcpClient = new TcpClient();
        tcpClient.Connect("127.0.0.1", 8888);

        Console.WriteLine("Enter messages to send (type 'exit' to quit):");

        while (true)
        {
            string message = Console.ReadLine();

            if (message.ToLower() == "exit")
                break;

            NetworkStream stream = tcpClient.GetStream();
            byte[] data = Encoding.UTF8.GetBytes(message);
            stream.Write(data, 0, data.Length);

            Console.WriteLine($"Sent: {message}");
        }

        tcpClient.Close();
    }
}
