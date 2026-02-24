const getJavaFunction = (url: string, dataStr: string) => {
    return `import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;

public class MessageSender {
    public static String sendMessage(String payload) throws IOException, InterruptedException {
        String url = "${url}";
        HttpClient client = HttpClient.newBuilder()
            .followRedirects(HttpClient.Redirect.NORMAL)
            .build();

        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(url))
            .POST(BodyPublishers.ofString(payload))
            .setHeader("Content-Type", "application/json")
            .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        return response.body();
    }

    public static void main(String[] args) throws Exception {
        String data = ${JSON.stringify(dataStr).slice(1, -1)};
        System.out.println(sendMessage(data));
    }
}`;
};

const getJavaScript = (url: string, dataStr: string) => {
    return `import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;

HttpClient client = HttpClient.newBuilder()
    .followRedirects(HttpClient.Redirect.NORMAL)
    .build();

HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${url}"))
    .POST(BodyPublishers.ofString(${JSON.stringify(dataStr).slice(1, -1)}))
    .setHeader("Content-Type", "application/json")
    .build();

HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());`;
};

export const getJava = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getJavaFunction(url, dataStr) : getJavaScript(url, dataStr);
};
