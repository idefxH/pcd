Name:           pcdp-wizard
Version:        0.3.7
Release:        1%{?dist}
Summary:        Interactive wizard for creating PCDP specifications
License:        GPL-2.0-only
URL:            https://example.com/pcdp-wizard
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.19
BuildRequires:  git

%description
pcdp-wizard is an interactive command-line tool for creating Post-Coding 
Development Paradigm (PCDP) specifications. It guides users through the 
process of creating complete, valid specification documents with session 
management and automatic validation.

%prep
%autosetup

%build
export CGO_ENABLED=0
go build -ldflags="-s -w" -o %{name} .

%install
install -D -m 755 %{name} %{buildroot}%{_bindir}/%{name}

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}

%changelog
* Mon Jan 15 2024 Matthias G. Eckermann <pcdp@mailbox.org> - 0.3.7-1
- Initial package