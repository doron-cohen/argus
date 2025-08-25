// Initialize API configuration first
import "./config/api";
// Initialize theme provider (sets data-theme if absent)
import "./ui/theme/theme-provider";

// Register all components explicitly
import "./components/component-details";
import "./components/component-list";
import "./pages/component-details";
import "./pages/home";
import "./pages/settings";
import "./router/outlet";

// Initialize the app last
import "./app";
