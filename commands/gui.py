import os
import sys
import threading
import tkinter as tk
from tkinter import messagebox
import click
import customtkinter as ctk
from tkinter import StringVar, Listbox
from rich.console import Console

from utils.github import get_available_versions, download_version
from utils.installer import (
    get_installed_versions, 
    install_from_zip, 
    uninstall_version, 
    is_version_installed
)
from utils.process import start_application, stop_application, get_process_status
from utils.config import get_version_config

console = Console()

# Custom listbox implementation for customtkinter
class CustomListbox(ctk.CTkFrame):
    def __init__(self, master, **kwargs):
        super().__init__(master, **kwargs)
        
        # Create a frame for the listbox
        self.listbox_frame = ctk.CTkFrame(self)
        self.listbox_frame.pack(fill="both", expand=True)
        
        # Create a scrollable frame
        self.scrollable_frame = ctk.CTkScrollableFrame(self.listbox_frame)
        self.scrollable_frame.pack(fill="both", expand=True)
        
        # List to store the items
        self.items = []
        self.selected_index = None
        self.callback = None
        
    def insert(self, index, item):
        """Insert an item at the specified index"""
        if index == "end":
            index = len(self.items)
            
        # Create a button for the item
        btn = ctk.CTkButton(
            self.scrollable_frame, 
            text=item,
            anchor="w",
            fg_color="transparent",
            text_color=("black", "white"),
            hover_color=("gray80", "gray20"),
            corner_radius=0,
            height=30,
            command=lambda i=len(self.items): self.select_item(i)
        )
        btn.pack(fill="x", padx=2, pady=1)
        
        # Add the item to the list
        self.items.append({"text": item, "button": btn})
    
    def delete(self, start, end=None):
        """Delete items from start to end"""
        if end == "end":
            end = len(self.items)
        elif end is None:
            end = start + 1
            
        # Remove the buttons
        for i in range(start, end):
            if i < len(self.items):
                self.items[i]["button"].destroy()
        
        # Remove the items from the list
        self.items = self.items[:start] + self.items[end:]
        self.selected_index = None
    
    def select_item(self, index):
        """Select an item"""
        # Deselect the previously selected item
        if self.selected_index is not None and self.selected_index < len(self.items):
            self.items[self.selected_index]["button"].configure(
                fg_color="transparent"
            )
        
        # Select the new item
        self.selected_index = index
        self.items[index]["button"].configure(
            fg_color=("gray70", "gray30")
        )
        
        # Call the callback
        if self.callback:
            self.callback(self.items[index]["text"])
    
    def bind(self, sequence, func, add=None):
        """Bind an event to the listbox"""
        # We only care about the command binding
        if sequence == "<<ListboxSelect>>":
            self.callback = func
    
    def get(self, index=None):
        """Get the selected item or item at index"""
        if index is not None:
            if 0 <= index < len(self.items):
                return self.items[index]["text"]
            return None
        
        if self.selected_index is not None and 0 <= self.selected_index < len(self.items):
            return self.items[self.selected_index]["text"]
        
        return None
    
    def configure(self, **kwargs):
        """Configure the listbox"""
        if "command" in kwargs:
            self.callback = kwargs["command"]
        
        super().configure(**kwargs)

class FgGui(ctk.CTk):
    def __init__(self):
        super().__init__()
        
        # Configure window
        self.title("FG - Java App Manager")
        self.geometry("800x600")
        
        # Set the appearance mode and color theme
        ctk.set_appearance_mode("System")
        ctk.set_default_color_theme("blue")
        
        # Create tabs
        self.tabview = ctk.CTkTabview(self)
        self.tabview.pack(padx=10, pady=10, fill="both", expand=True)
        
        # Add tabs
        self.tab_versions = self.tabview.add("Versions")
        self.tab_running = self.tabview.add("Running")
        
        # Set up the versions tab
        self.setup_versions_tab()
        
        # Set up the running tab
        self.setup_running_tab()
        
        # Initialize data
        self.available_versions = []
        self.installed_versions = []
        self.running_processes = []
        
        # Status message label
        self.status_text = None
        
        # Initial data load
        self.refresh_data()
    
    def setup_versions_tab(self):
        """Set up the versions tab UI"""
        # Create frames
        frame_left = ctk.CTkFrame(self.tab_versions, width=300)
        frame_left.pack(side="left", fill="both", padx=10, pady=10)
        
        frame_right = ctk.CTkFrame(self.tab_versions)
        frame_right.pack(side="right", fill="both", expand=True, padx=10, pady=10)
        
        # Available versions list (left side)
        ctk.CTkLabel(frame_left, text="Available Versions", font=("Arial", 16, "bold")).pack(pady=5)
        
        self.available_listbox = CustomListbox(frame_left)
        self.available_listbox.pack(fill="both", expand=True, padx=10, pady=10)
        
        btn_install = ctk.CTkButton(frame_left, text="Install Selected", command=self.install_selected)
        btn_install.pack(padx=10, pady=10, fill="x")
        
        btn_refresh = ctk.CTkButton(frame_left, text="Refresh", command=self.refresh_data)
        btn_refresh.pack(padx=10, pady=10, fill="x")
        
        # Installed versions list (right side)
        ctk.CTkLabel(frame_right, text="Installed Versions", font=("Arial", 16, "bold")).pack(pady=5)
        
        self.installed_listbox = CustomListbox(frame_right)
        self.installed_listbox.pack(fill="both", expand=True, padx=10, pady=10)
        
        # Buttons for installed versions
        btn_frame = ctk.CTkFrame(frame_right)
        btn_frame.pack(fill="x", padx=10, pady=10)
        
        btn_start = ctk.CTkButton(btn_frame, text="Start", command=self.start_selected)
        btn_start.pack(side="left", padx=5, expand=True, fill="x")
        
        btn_uninstall = ctk.CTkButton(btn_frame, text="Uninstall", command=self.uninstall_selected)
        btn_uninstall.pack(side="right", padx=5, expand=True, fill="x")
    
    def setup_running_tab(self):
        """Set up the running instances tab UI"""
        # Create frame for running processes
        frame = ctk.CTkFrame(self.tab_running)
        frame.pack(fill="both", expand=True, padx=10, pady=10)
        
        # Running processes list
        ctk.CTkLabel(frame, text="Running Instances", font=("Arial", 16, "bold")).pack(pady=5)
        
        self.running_listbox = CustomListbox(frame)
        self.running_listbox.pack(fill="both", expand=True, padx=10, pady=10)
        
        # Buttons for running instances
        btn_frame = ctk.CTkFrame(frame)
        btn_frame.pack(fill="x", padx=10, pady=10)
        
        btn_stop = ctk.CTkButton(btn_frame, text="Stop Selected", command=self.stop_selected)
        btn_stop.pack(side="left", padx=5, expand=True, fill="x")
        
        btn_refresh_running = ctk.CTkButton(btn_frame, text="Refresh", command=self.refresh_running)
        btn_refresh_running.pack(side="right", padx=5, expand=True, fill="x")
    
    def set_status(self, message):
        """Set or update status message"""
        # Remove existing status if any
        self.clear_status()
        
        # Create new status label
        self.status_text = ctk.CTkLabel(self, text=message, font=("Arial", 12))
        self.status_text.pack(side="bottom", fill="x", padx=10, pady=5)
        self.update_idletasks()
    
    def clear_status(self):
        """Clear the status message"""
        if self.status_text is not None:
            self.status_text.destroy()
            self.status_text = None
            self.update_idletasks()
    
    def refresh_data(self):
        """Refresh data from GitHub and local installations"""
        # Show loading status
        self.set_status("Loading data...")
        
        # Start a thread to fetch data
        threading.Thread(target=self._fetch_data, daemon=True).start()
    
    def _fetch_data(self):
        """Fetch data in a background thread"""
        # Get available versions
        self.available_versions = get_available_versions()
        
        # Get installed versions
        self.installed_versions = get_installed_versions()
        
        # Get running processes
        self.running_processes = get_process_status()
        
        # Update UI
        self.after(100, self._update_ui_with_data)
    
    def _update_ui_with_data(self):
        """Update the UI with fetched data"""
        # Update available versions listbox
        self.available_listbox.delete(0, "end")
        for version_info in self.available_versions:
            self.available_listbox.insert("end", version_info['version'])
        
        # Update installed versions listbox
        self.installed_listbox.delete(0, "end")
        for version in self.installed_versions:
            self.installed_listbox.insert("end", version)
        
        # Update running processes listbox
        self.refresh_running()
        
        # Clear status
        self.clear_status()
    
    def refresh_running(self):
        """Refresh the list of running processes"""
        self.running_processes = get_process_status()
        self.running_listbox.delete(0, "end")
        
        for proc in self.running_processes:
            status = "Running" if proc.get('running', False) else "Stopped"
            display_text = f"PID: {proc['pid']} - Version: {proc['version']} - {status}"
            self.running_listbox.insert("end", display_text)
    
    def on_available_select(self, selected_item):
        """Handle selection of an available version"""
        pass
    
    def on_installed_select(self, selected_item):
        """Handle selection of an installed version"""
        pass
    
    def on_running_select(self, selected_item):
        """Handle selection of a running process"""
        pass
    
    def install_selected(self):
        """Install the selected version"""
        selection = self.available_listbox.get()
        if not selection:
            self.show_message("Error", "Please select a version to install")
            return
        
        # Find the version in available_versions
        version = selection
        
        # Check if already installed
        if is_version_installed(version):
            if not self.ask_confirmation(f"Version {version} is already installed. Reinstall?"):
                return
        
        # Download and install
        self.set_status(f"Installing {version}...")
        
        # Install in a background thread
        threading.Thread(target=self._install_version, args=(version,), daemon=True).start()
    
    def _install_version(self, version):
        """Install a version in a background thread"""
        zip_path = download_version(version)
        
        if not zip_path:
            self.after(100, lambda: self._handle_failed_download(version))
            return
        
        success = install_from_zip(zip_path, version)
        
        # Update UI
        self.after(100, lambda: self._finish_installation(version, success))
    
    def _handle_failed_download(self, version):
        """Handle failed download"""
        self.clear_status()
        self.show_message("Error", f"Failed to download version {version}")
    
    def _finish_installation(self, version, success):
        """Handle installation completion"""
        # Clear the status message first
        self.clear_status()
        
        # Show success/error message
        if success:
            self.show_message("Success", f"Version {version} installed successfully")
        else:
            self.show_message("Error", f"Failed to install version {version}")
        
        # Refresh data
        self.refresh_data()
    
    def start_selected(self):
        """Start the selected version"""
        selection = self.installed_listbox.get()
        if not selection:
            self.show_message("Error", "Please select a version to start")
            return
        
        # Start the application
        pid = start_application(selection)
        
        if pid:
            self.show_message("Success", f"Application started successfully. PID: {pid}")
            self.refresh_running()
        else:
            self.show_message("Error", "Failed to start the application")
    
    def stop_selected(self):
        """Stop the selected process"""
        selection = self.running_listbox.get()
        if not selection:
            self.show_message("Error", "Please select a process to stop")
            return
        
        # Extract PID from the selection text
        pid = selection.split(" - ")[0].replace("PID: ", "")
        
        # Stop the process
        success = stop_application(pid)
        
        if success:
            self.show_message("Success", f"Application (PID: {pid}) stopped successfully")
            self.refresh_running()
        else:
            self.show_message("Error", f"Failed to stop process {pid}")
    
    def uninstall_selected(self):
        """Uninstall the selected version"""
        selection = self.installed_listbox.get()
        if not selection:
            self.show_message("Error", "Please select a version to uninstall")
            return
        
        # Confirm uninstallation
        if not self.ask_confirmation(f"Are you sure you want to uninstall version {selection}?"):
            return
        
        # Uninstall the version
        success = uninstall_version(selection)
        
        if success:
            self.show_message("Success", f"Version {selection} uninstalled successfully")
            self.refresh_data()
        else:
            self.show_message("Error", f"Failed to uninstall version {selection}")
    
    def show_message(self, title, message):
        """Show a message dialog"""
        # Using standard messagebox from tkinter instead of CTkInputDialog
        messagebox.showinfo(title, message)
    
    def ask_confirmation(self, message):
        """Ask for confirmation"""
        # Using standard messagebox from tkinter for confirmation
        return messagebox.askyesno("Confirmation", message)

@click.command()
def gui():
    """Launch the graphical user interface."""
    # Initialize the GUI
    app = FgGui()
    app.mainloop() 