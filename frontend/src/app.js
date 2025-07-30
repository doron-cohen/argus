import Alpine from "alpinejs";
import ComponentsList from "./components/components-list.js";
import "./styles/app.css";

// Register Alpine.js components
Alpine.data("componentsList", ComponentsList);

// Start Alpine.js
Alpine.start();
