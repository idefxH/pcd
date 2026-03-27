Name:           pcdp-templates
Version:        0.3.19
Release:        1
Summary:        Deployment templates and library hints for the Post-Coding Development Paradigm
BuildArch:      noarch

License:        CC-BY-4.0
URL:            https://github.com/mge1512/pcdp

Source0:        pcdp-templates-%{version}.tar.gz

# No build dependencies — data package only
BuildRequires:  (nothing)

# Both tools require this package at runtime
# They are not listed here as Requires — this is a data package,
# the tools declare Requires: pcdp-templates in their own spec files.

%description
pcdp-templates provides the deployment templates and library hints files
for the Post-Coding Development Paradigm (PCDP).

Deployment templates define language defaults, packaging conventions,
and AI translation execution recipes for each supported deployment type.
Library hints files provide verified dependency versions and API shapes
for common libraries.

Both pcdp-lint and mcp-server-pcdp read from the installed template
and hints directories at runtime.

Templates are installed under:
  /usr/share/pcdp/templates/

Hints are installed under:
  /usr/share/pcdp/hints/

%prep
%setup -q

%build
# Nothing to build — data package

%install
install -d %{buildroot}%{_datadir}/pcdp/templates
install -d %{buildroot}%{_datadir}/pcdp/hints

# Templates
install -m 0644 templates/backend-service.template.md  %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/cli-tool.template.md         %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/cloud-native.template.md     %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/gui-tool.template.md         %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/library-c-abi.template.md    %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/mcp-server.template.md       %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/project-manifest.template.md %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/python-tool.template.md      %{buildroot}%{_datadir}/pcdp/templates/
install -m 0644 templates/verified-library.template.md %{buildroot}%{_datadir}/pcdp/templates/

# Hints
install -m 0644 hints/cloud-native.go.go-libvirt.hints.md       %{buildroot}%{_datadir}/pcdp/hints/
install -m 0644 hints/cloud-native.go.golang-crypto-ssh.hints.md %{buildroot}%{_datadir}/pcdp/hints/
install -m 0644 hints/mcp-server.go.mcp-go.hints.md             %{buildroot}%{_datadir}/pcdp/hints/

%files
%license LICENSE
%dir %{_datadir}/pcdp
%dir %{_datadir}/pcdp/templates
%dir %{_datadir}/pcdp/hints

# Templates
%{_datadir}/pcdp/templates/backend-service.template.md
%{_datadir}/pcdp/templates/cli-tool.template.md
%{_datadir}/pcdp/templates/cloud-native.template.md
%{_datadir}/pcdp/templates/gui-tool.template.md
%{_datadir}/pcdp/templates/library-c-abi.template.md
%{_datadir}/pcdp/templates/mcp-server.template.md
%{_datadir}/pcdp/templates/project-manifest.template.md
%{_datadir}/pcdp/templates/python-tool.template.md
%{_datadir}/pcdp/templates/verified-library.template.md

# Hints
%{_datadir}/pcdp/hints/cloud-native.go.go-libvirt.hints.md
%{_datadir}/pcdp/hints/cloud-native.go.golang-crypto-ssh.hints.md
%{_datadir}/pcdp/hints/mcp-server.go.mcp-go.hints.md

%changelog
* Fri Mar 27 2026 Matthias G. Eckermann <pcdp@mailbox.org> - 0.3.19-1
- Initial release: all deployment templates and library hints
