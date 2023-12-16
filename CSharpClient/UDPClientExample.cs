using System.Net;
using System.Net.Sockets;
using System.Text;

class UDPClient
{
    private UdpClient? udpClient;

    public void Start()
    {
        udpClient = new UdpClient();
        udpClient.Connect("127.0.0.1", 8889);

        // Start a thread for receiving messages
        Thread receiveThread = new Thread(ReceiveMessages);
        receiveThread.Start();

        Console.WriteLine("Enter messages to send (type 'exit' to quit):");

        while (true)
        {
            string? message = Console.ReadLine();

            if (message == null)
                break;

            if (message.ToLower() == "exit")
                break;

            byte[] data = Encoding.UTF8.GetBytes(message);
            udpClient.Send(data, data.Length);

            Console.WriteLine($"Sent: {message}");
        }

        // Close the UDP client when done
        udpClient.Close();
    }

    private void ReceiveMessages()
    {
        if (udpClient == null)
            return;

        try
        {
            IPEndPoint remoteEndpoint = new IPEndPoint(IPAddress.Any, 0);

            while (true)
            {
                byte[] receivedData = udpClient.Receive(ref remoteEndpoint);
                string receivedMessage = Encoding.UTF8.GetString(receivedData);

                Console.WriteLine($"Received: {receivedMessage}");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error receiving messages: {ex.Message}");
        }
    }
}
