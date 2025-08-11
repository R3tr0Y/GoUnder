// script.js

function analyzeWebsite() {
  const website = document.getElementById("website").value.trim();
  if (!website) {
    alert("请输入网址！");
    return;
  }

  // 清空之前的结果
  document.getElementById("real-ip").innerText = "正在加载...";
  document.getElementById("cloud-service").innerText = "正在加载...";
  document.getElementById("website-architecture").innerText = "正在加载...";
  document.getElementById("middleware").innerText = "正在加载...";

  // 动态获取当前站点的根地址
  const baseURL = `${window.location.origin}/api/analyze?website=${encodeURIComponent(website)}`;

  axios.get(baseURL)
    .then(response => {
      const data = response.data;

      document.getElementById("real-ip").innerText = `真实 IP: ${data.ip ?? "未知"}`;
      document.getElementById("cloud-service").innerText = `云服务: ${data.cloudService ?? "未知"}`;
      document.getElementById("website-architecture").innerText = `架构: ${data.architecture ?? "未知"}`;
      document.getElementById("middleware").innerText = `中间件: ${data.middleware ?? "未知"}`;
    })
    .catch(error => {
      console.error("分析失败:", error);
      alert("分析失败，请稍后再试！");
    });
}

