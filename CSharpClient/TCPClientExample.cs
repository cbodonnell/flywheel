using System.Net.Sockets;
using System.Text;

class TCPClient
{
    private TcpClient? tcpClient;
    private Thread? receiveThread;

    public void Start()
    {
        tcpClient = new TcpClient();
        tcpClient.Connect("127.0.0.1", 8888);

        // Start a thread for receiving messages
        receiveThread = new Thread(ReceiveMessages);
        receiveThread.Start();

        Console.WriteLine("Enter messages to send (type 'exit' to quit):");

        while (true)
        {
            string? message = Console.ReadLine();
            if (message == null)
                break;

            if (message.ToLower() == "exit")
                break;

            NetworkStream stream = tcpClient.GetStream();
            byte[] data = Encoding.UTF8.GetBytes(message);
            stream.Write(data, 0, data.Length);

            Console.WriteLine($"Sent: {message}");
        }

        // Close the TCP client and stop the receive thread when done
        tcpClient.Close();
        receiveThread.Join();
    }

    private void ReceiveMessages()
    {
        if (tcpClient == null)
            return;

        try
        {
            NetworkStream stream = tcpClient.GetStream();
            byte[] buffer = new byte[1024];

            while (true)
            {
                int bytesRead = stream.Read(buffer, 0, buffer.Length);
                if (bytesRead == 0)
                    break;

                string receivedMessage = Encoding.UTF8.GetString(buffer, 0, bytesRead);
                Console.WriteLine($"Received: {receivedMessage}");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error receiving messages: {ex.Message}");
        }
    }
}
