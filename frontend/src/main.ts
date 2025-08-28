// Import global styles first
import "./styles.css";

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

// Import new UI components
import "./ui/components/ui-empty-state";
import "./ui/components/ui-loading-indicator";
import "./ui/components/ui-description-list";

// Initialize the app last
import "./app";
