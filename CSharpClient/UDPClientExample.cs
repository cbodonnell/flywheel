// using System;
// using System.Net;
// using System.Net.Sockets;
// using System.Text;

// class Program
// {
//     static void Main()
//     {
//         UdpClient udpClient = new UdpClient();
//         udpClient.Connect("127.0.0.1", 8889);

//         Console.WriteLine("Enter messages to send (type 'exit' to quit):");

//         while (true)
//         {
//             string message = Console.ReadLine();

//             if (message.ToLower() == "exit")
//                 break;

//             byte[] data = Encoding.UTF8.GetBytes(message);
//             udpClient.Send(data, data.Length);

//             Console.WriteLine($"Sent: {message}");
//         }

//         udpClient.Close();
//     }
// }
