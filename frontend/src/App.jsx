// import { useEffect, useState } from "react";
// import "./App.css";

// function App() {
//   const [status, setStatus] = useState({
//   circuit: "",
//   activeRoute: "",
//   requests: 0,
//   rps: 0,
// });

//   useEffect(() => {
//   const socket = new WebSocket("ws://localhost:8080/ws");

//   socket.onopen = () => {
//     console.log("WebSocket connected");
//   };

//   socket.onmessage = (event) => {
//     try {
//       const data = JSON.parse(event.data);

//       setStatus(data);
//     } catch (error) {
//       console.error(
//         "WebSocket message parsing error:",
//         error
//       );
//     }
//   };

//   socket.onerror = (error) => {
//     console.error(
//       "WebSocket error:",
//       error
//     );
//   };

//   socket.onclose = () => {
//     console.log(
//       "WebSocket disconnected"
//     );
//   };

//   return () => {
//     socket.close();
//   };
// }, []);

//   return (
//     <div className="container">
//       <h1>👨‍💻 Multiuser Dashboard</h1>

//       <div className="cards">

//         <div className="card">
//           <h2>Circuit Breaker</h2>
//           <h3>{status.circuit}</h3>
//         </div>

//         <div className="card">
//           <h2>Active Route</h2>
//           <h3>{status.activeRoute}</h3>
//         </div>

//         <div className="card">
//           <h2>Requests / Second</h2>
//           <h3>{status.rps}</h3>
//         </div>

//         <div className="card">
//           <h2>Total Requests</h2>
//           <h3>{status.requests}</h3>
//         </div>

//       </div>
//     </div>
//   );
// }

// export default App;

// import { useEffect, useRef, useState } from "react";
// import "./App.css";

// function App() {
//   const [status, setStatus] = useState({
//     circuit: "",
//     activeRoute: "",
//     requests: 0,
//     rps: 0,
//   });

//   // Store the latest WebSocket data without
//   // triggering React rendering for every message.
//   const latestStatus = useRef(status);

//   // Store the animation-frame ID.
//   const frameRef = useRef(null);

//   useEffect(() => {
//     const socket = new WebSocket("ws://localhost:8080/ws");

//     socket.onopen = () => {
//       console.log("WebSocket connected");
//     };

//     socket.onmessage = (event) => {
//       try {
//         const data = JSON.parse(event.data);

//         // Keep only the latest message.
//         latestStatus.current = data;

//         // Update React at most once per browser frame.
//         if (frameRef.current === null) {
//           frameRef.current = requestAnimationFrame(() => {
//             setStatus(latestStatus.current);
//             frameRef.current = null;
//           });
//         }
//       } catch (error) {
//         console.error("WebSocket message parsing error:", error);
//       }
//     };

//     socket.onerror = (error) => {
//       console.error("WebSocket error:", error);
//     };

//     socket.onclose = () => {
//       console.log("WebSocket disconnected");
//     };

//     return () => {
//       socket.close();

//       if (frameRef.current !== null) {
//         cancelAnimationFrame(frameRef.current);
//         frameRef.current = null;
//       }
//     };
//   }, []);

//   const circuitClass = status.circuit.toLowerCase();
//   const routeClass = status.activeRoute.toLowerCase();

//   return (
//     <div className="container">
//       <h1>👨‍💻 Multiuser Dashboard</h1>

//       <div className="status-banner">
//         <div className={`circuit-indicator ${circuitClass}`}>
//           <span className="indicator-dot"></span>

//           <div>
//             <strong>Circuit Breaker</strong>
//             <span>{status.circuit || "UNKNOWN"}</span>
//           </div>
//         </div>

//         <div className={`route-indicator ${routeClass}`}>
//           <strong>Active Route</strong>
//           <span>{status.activeRoute || "UNKNOWN"}</span>
//         </div>
//       </div>

//       <div className="cards">
//         <div className="card">
//           <h2>Circuit Breaker</h2>
//           <h3>{status.circuit}</h3>
//         </div>

//         <div className="card">
//           <h2>Active Route</h2>
//           <h3>{status.activeRoute}</h3>
//         </div>

//         <div className="card">
//           <h2>Requests / Second</h2>
//           <h3>{status.rps}</h3>
//         </div>

//         <div className="card">
//           <h2>Total Requests</h2>
//           <h3>{status.requests}</h3>
//         </div>
//       </div>
//     </div>
//   );
// }

// export default App;

// import { useEffect, useState } from "react";
// import "./App.css";

// function App() {
//   const [status, setStatus] = useState({
//     circuit: "",
//     activeRoute: "",
//     requests: 0,
//     rps: 0,
//   });

//   useEffect(() => {
//     const socket = new WebSocket("ws://localhost:8080/ws");

//     socket.onopen = () => {
//       console.log("WebSocket connected");
//     };

//     socket.onmessage = (event) => {
//       try {
//         const data = JSON.parse(event.data);
//         setStatus(data);
//       } catch (error) {
//         console.error("WebSocket message parsing error:", error);
//       }
//     };

//     socket.onerror = (error) => {
//       console.error("WebSocket error:", error);
//     };

//     socket.onclose = () => {
//       console.log("WebSocket disconnected");
//     };

//     return () => {
//       socket.close();
//     };
//   }, []);

//   const circuitClass = status.circuit
//     ? status.circuit.toLowerCase().replace("-", "-")
//     : "";

//   const routeClass = status.activeRoute
//     ? status.activeRoute.toLowerCase()
//     : "";

//   return (
//     <div className="container">

//       <h1>👨‍💻 Multiuser Dashboard</h1>

//       {/* =====================================================
//           STATUS INDICATORS
//       ====================================================== */}

//       <div className="status-banner">

//         <div className={`circuit-indicator ${circuitClass}`}>
//           <span className="indicator-dot"></span>

//           <div>
//             <strong>Circuit Breaker</strong>
//             <span>{status.circuit}</span>
//           </div>
//         </div>

//         <div className={`route-indicator ${routeClass}`}>
//           <strong>Active Route</strong>
//           <span>{status.activeRoute}</span>
//         </div>

//       </div>

//       {/* =====================================================
//           REQUEST STATISTICS
//       ====================================================== */}

//       <div className="stats">

//         <div className="card">
//           <h2>Requests / Second</h2>
//           <h3>{status.rps}</h3>
//         </div>

//         <div className="card">
//           <h2>Total Requests</h2>
//           <h3>{status.requests}</h3>
//         </div>

//       </div>

//     </div>
//   );
// }

// export default App;

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
    const socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
      console.log("WebSocket connected");
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setStatus(data);
      } catch (error) {
        console.error("WebSocket message parsing error:", error);
      }
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    socket.onclose = () => {
      console.log("WebSocket disconnected");
    };

    return () => {
      socket.close();
    };
  }, []);

  const circuitClass = status.circuit
    ? status.circuit.toLowerCase()
    : "";

  const routeClass = status.activeRoute
    ? status.activeRoute.toLowerCase()
    : "";

  return (
    <div className="container">

      <h1>👨‍💻 Multiuser Dashboard</h1>

      {/* =====================================================
          CIRCUIT BREAKER + ACTIVE ROUTE
      ====================================================== */}

      <div className="status-banner">

        <div className={`circuit-indicator ${circuitClass}`}>

          <span className="indicator-dot"></span>

          <div>
            <strong>Circuit Breaker</strong>
            <span>{status.circuit}</span>
          </div>

        </div>

        <div className={`route-indicator ${routeClass}`}>

          <strong>Active Route</strong>
          <span>{status.activeRoute}</span>

        </div>

      </div>

      {/* =====================================================
          REQUEST STATISTICS
      ====================================================== */}

      <div className="stats">

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