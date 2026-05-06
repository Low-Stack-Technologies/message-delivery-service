/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import {
  createAdminEmailAccount,
  createAdminMessage,
  createAdminService,
  deleteAdminEmailAccount,
  deleteAdminService,
  getAdminDashboard,
  getAdminSmsCredentials,
  listAdminEmailAccounts,
  listAdminMessages,
  listAdminServices,
  rerollAdminService,
  rotateAdminSmsCredentials,
  testAdminEmailAccount,
  updateAdminEmailAccount,
  updateAdminService,
  updateAdminSmsCredentials,
  type AdminActivity,
  type AdminEmailAccount,
  type AdminMessage,
  type AdminService,
  type AdminSmsCredentials,
} from "./admin-api";
import { useAdminAuth } from "./auth";

export type ServiceScope = "all" | "email" | "sms";
export type ServiceStatus = "active" | "paused";
export type EmailAccountStatus = "healthy" | "warning" | "offline";
export type ActivityTone = "info" | "success" | "warning" | "danger";

export type ServiceRecord = AdminService;
export type EmailAccountRecord = AdminEmailAccount;
export type SmsCredentialRecord = AdminSmsCredentials;
export type SentMessageRecord = AdminMessage;
export type ActivityRecord = AdminActivity;

type AdminState = {
  services: ServiceRecord[];
  emailAccounts: EmailAccountRecord[];
  smsCredentials: SmsCredentialRecord;
  messages: SentMessageRecord[];
  activity: ActivityRecord[];
};

type AdminStore = {
  state: AdminState;
  dispatch: (action: Action) => Promise<void>;
  refresh: () => Promise<void>;
};

type Action =
  | {
      type: "service/add";
      payload: Omit<ServiceRecord, "createdAt" | "lastRerollAt">;
    }
  | {
      type: "service/delete";
      payload: { id: string };
    }
  | {
      type: "service/reroll";
      payload: { id: string };
    }
  | {
      type: "service/set-status";
      payload: { id: string; status: ServiceStatus };
    }
  | {
      type: "service/set-scope";
      payload: { id: string; scope: ServiceScope };
    }
  | {
      type: "email/add";
      payload: Omit<
        EmailAccountRecord,
        "isDefault" | "status" | "lastTestedAt"
      >;
    }
  | {
      type: "email/delete";
      payload: { id: string };
    }
  | {
      type: "email/set-default";
      payload: { id: string };
    }
  | {
      type: "email/test";
      payload: { id: string };
    }
  | {
      type: "sms/update";
      payload: Pick<SmsCredentialRecord, "username" | "password">;
    }
  | {
      type: "sms/rotate";
    }
  | {
      type: "message/send";
      payload: Omit<SentMessageRecord, "id" | "createdAt" | "status"> & {
        templateData: string;
      };
    };

const emptyState: AdminState = {
  services: [],
  emailAccounts: [],
  smsCredentials: {
    username: "",
    password: "",
    status: "stale",
    lastSyncedAt: "",
    rotationCount: 0,
  },
  messages: [],
  activity: [],
};

const AdminStoreContext = createContext<AdminStore | null>(null);

function parseTemplateData(value: string) {
  if (!value.trim()) {
    return {};
  }

  try {
    return JSON.parse(value) as Record<string, unknown>;
  } catch {
    return { raw: value };
  }
}

export function AdminStoreProvider({ children }: { children: ReactNode }) {
  const { token } = useAdminAuth();
  const [state, setState] = useState<AdminState>(emptyState);

  async function refresh() {
    if (!token) {
      setState(emptyState);
      return;
    }

    const [dashboard, services, emailAccounts, smsCredentials, messages] =
      await Promise.all([
        getAdminDashboard(token),
        listAdminServices(token),
        listAdminEmailAccounts(token),
        getAdminSmsCredentials(token),
        listAdminMessages(token, { limit: 100 }),
      ]);

    setState({
      services: services.services,
      emailAccounts: emailAccounts.emailAccounts,
      smsCredentials: smsCredentials.smsCredentials,
      messages: messages.messages,
      activity: dashboard.recentActivity,
    });
  }

  useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!token) {
        if (!cancelled) {
          setState(emptyState);
        }
        return;
      }

      try {
        await refresh();
      } catch (error) {
        if (!cancelled) {
          console.error("Failed to load admin data", error);
          setState(emptyState);
        }
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
    // refresh depends on token through closure, so token is the only trigger we need.
  }, [token]);

  async function dispatch(action: Action) {
    if (!token) {
      return;
    }

    try {
      switch (action.type) {
        case "service/add":
          await createAdminService(token, {
            ...action.payload,
            owner: action.payload.owner,
          });
          break;
        case "service/delete":
          await deleteAdminService(token, action.payload.id);
          break;
        case "service/reroll":
          await rerollAdminService(token, action.payload.id);
          break;
        case "service/set-status":
          await updateAdminService(token, action.payload.id, {
            status: action.payload.status,
          });
          break;
        case "service/set-scope":
          await updateAdminService(token, action.payload.id, {
            scope: action.payload.scope,
          });
          break;
        case "email/add":
          await createAdminEmailAccount(token, {
            ...action.payload,
            isDefault: state.emailAccounts.length === 0,
          });
          break;
        case "email/delete":
          await deleteAdminEmailAccount(token, action.payload.id);
          break;
        case "email/set-default":
          await updateAdminEmailAccount(token, action.payload.id, {
            isDefault: true,
          });
          break;
        case "email/test":
          await testAdminEmailAccount(token, action.payload.id);
          break;
        case "sms/update":
          await updateAdminSmsCredentials(token, action.payload);
          break;
        case "sms/rotate":
          await rotateAdminSmsCredentials(token);
          break;
        case "message/send":
          await createAdminMessage(token, {
            channel: action.payload.channel,
            serviceId: action.payload.serviceId,
            recipients: action.payload.recipients,
            from:
              action.payload.channel === "email"
                ? action.payload.sender
                : undefined,
            senderName:
              action.payload.channel === "sms"
                ? action.payload.sender
                : undefined,
            subject:
              action.payload.channel === "email"
                ? (action.payload.subject ?? undefined)
                : undefined,
            contentMode: action.payload.contentMode,
            body:
              action.payload.contentMode === "template"
                ? undefined
                : (action.payload.body ?? undefined),
            template:
              action.payload.contentMode === "template"
                ? {
                    name: action.payload.templateName ?? "template",
                    data: parseTemplateData(action.payload.templateData),
                  }
                : undefined,
          });
          break;
      }

      await refresh();
    } catch (error) {
      console.error("Admin mutation failed", error);
    }
  }

  return (
    <AdminStoreContext.Provider value={{ state, dispatch, refresh }}>
      {children}
    </AdminStoreContext.Provider>
  );
}

export function useAdminStore() {
  const value = useContext(AdminStoreContext);

  if (!value) {
    throw new Error("useAdminStore must be used inside AdminStoreProvider");
  }

  return value;
}
