import { useEffect, useState } from "react";
import "./App.css";

function App() {
  const [status, setStatus] = useState({
  circuit: "",
  activeRoute: "",
  requests: 0,
  rps: 0,
});

  useEffect(() => {
    const fetchStatus = () => {
      fetch("http://localhost:8080/status")
        .then((res) => res.json())
        .then((data) => setStatus(data))
        .catch((err) => console.error(err));
    };

    fetchStatus();

    const timer = setInterval(fetchStatus, 200);

    return () => clearInterval(timer);
  }, []);

  return (
    <div className="container">
      <h1>👨‍💻 Multiuser Dashboard</h1>

      <div className="cards">

        <div className="card">
          <h2>Circuit Breaker</h2>
          <h3>{status.circuit}</h3>
        </div>

        <div className="card">
          <h2>Active Route</h2>
          <h3>{status.activeRoute}</h3>
        </div>

        <div className="card">
          <h2>Requests / Second</h2>
          <h3>{status.rps}</h3>
        </div>

        <div className="card">
          <h2>Total Requests</h2>
          <h3>{status.requests}</h3>
        </div>

      </div>
    </div>
  );
}

export default App;